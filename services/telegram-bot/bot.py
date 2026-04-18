import asyncio
import platform

if platform.system() == "Darwin":
    asyncio.set_event_loop_policy(asyncio.DefaultEventLoopPolicy())
import logging
import re
import sqlite3
from typing import Optional, Dict, Any

import requests
from aiogram import Bot, Dispatcher, types
from aiogram.utils import executor
from aiogram.types import ReplyKeyboardMarkup, KeyboardButton, InlineKeyboardMarkup, InlineKeyboardButton

from config import BOT_TOKEN, ADMIN_CHAT_ID, SITE_API_URL, DB_PATH

logging.basicConfig(level=logging.INFO)

bot = Bot(token=BOT_TOKEN)
dp = Dispatcher(bot)

# -----------------------------
# Локальное хранилище состояния (SQLite)
# 1) согласие на уведомления
# 2) последний отправленный notification_id
# -----------------------------

def db_conn():
    conn = sqlite3.connect(DB_PATH)
    conn.row_factory = sqlite3.Row
    return conn

def init_state_db():
    conn = db_conn()
    cur = conn.cursor()
    cur.execute("""
        CREATE TABLE IF NOT EXISTS user_state (
            tg_user_id INTEGER PRIMARY KEY,
            notify_enabled INTEGER DEFAULT 0
        )
    """)
    cur.execute("""
        CREATE TABLE IF NOT EXISTS bot_state (
            key TEXT PRIMARY KEY,
            value TEXT
        )
    """)
    conn.commit()
    conn.close()

def set_notify_enabled(tg_user_id: int, enabled: bool):
    conn = db_conn()
    cur = conn.cursor()
    cur.execute("""
        INSERT INTO user_state (tg_user_id, notify_enabled)
        VALUES (?, ?)
        ON CONFLICT(tg_user_id) DO UPDATE SET notify_enabled=excluded.notify_enabled
    """, (tg_user_id, 1 if enabled else 0))
    conn.commit()
    conn.close()

def get_notify_enabled(tg_user_id: int) -> bool:
    conn = db_conn()
    cur = conn.cursor()
    cur.execute("SELECT notify_enabled FROM user_state WHERE tg_user_id=?", (tg_user_id,))
    row = cur.fetchone()
    conn.close()
    return bool(row["notify_enabled"]) if row else False

def set_bot_state(key: str, value: str):
    conn = db_conn()
    cur = conn.cursor()
    cur.execute("""
        INSERT INTO bot_state (key, value)
        VALUES (?, ?)
        ON CONFLICT(key) DO UPDATE SET value=excluded.value
    """, (key, value))
    conn.commit()
    conn.close()

def get_bot_state(key: str) -> Optional[str]:
    conn = db_conn()
    cur = conn.cursor()
    cur.execute("SELECT value FROM bot_state WHERE key=?", (key,))
    row = cur.fetchone()
    conn.close()
    return row["value"] if row else None

init_state_db()

# -----------------------------
# Нормализация данных
# -----------------------------

def normalize_username(u: str) -> str:
    return (u or "").lstrip("@").strip()

def normalize_phone(raw: str) -> str:
    """
    В вашем бэке phone_number валидируется regex '^8\\d{10}$' (см. schemas/users.py).
    Поэтому приводим к 8XXXXXXXXXX, если пришло +7XXXXXXXXXX или 7XXXXXXXXXX.
    """
    if not raw:
        return ""
    digits = re.sub(r"\D+", "", raw)

    if len(digits) == 11 and digits.startswith("7"):
        return "8" + digits[1:]
    if len(digits) == 11 and digits.startswith("8"):
        return digits
    return digits

# -----------------------------
# Клиент к вашему API
# -----------------------------

def api_get_user_profile(username: str) -> Optional[Dict[str, Any]]:
    """
    GET /users/profile/{username}
    """
    try:
        url = f"{SITE_API_URL}/users/profile/{username}"
        r = requests.get(url, timeout=10)
        if r.status_code == 200:
            return r.json()
        return None
    except Exception:
        logging.exception("api_get_user_profile failed")
        return None

def api_patch_user(user_id: int, payload: Dict[str, Any]) -> bool:
    """
    PATCH /users/{user_id}
    """
    try:
        url = f"{SITE_API_URL}/users/{user_id}"
        r = requests.patch(url, json=payload, timeout=10)
        return r.status_code in (200, 201)
    except Exception:
        logging.exception("api_patch_user failed")
        return False

def api_get_notifications(limit: int = 50, offset: int = 0) -> Optional[Dict[str, Any]]:
    """
    GET /notification/all?limit=..&offset=..
    Возвращает { items, total, limit, offset, has_more }
    """
    try:
        url = f"{SITE_API_URL}/notification/all"
        r = requests.get(url, params={"limit": limit, "offset": offset}, timeout=10)
        if r.status_code == 200:
            return r.json()
        return None
    except Exception:
        logging.exception("api_get_notifications failed")
        return None

# -----------------------------
# UI /start + контакт
# -----------------------------

@dp.message_handler(commands=['start'])
async def cmd_start(message: types.Message):
    kb = ReplyKeyboardMarkup(resize_keyboard=True)
    kb.add(KeyboardButton(text='📱 Поделиться контактом', request_contact=True))
    await message.answer(
        "👋 Привет! Чтобы продолжить, поделись своим контактом:",
        reply_markup=kb
    )

@dp.message_handler(content_types=['contact'])
async def handle_contact(message: types.Message):
    tg_user_id = message.from_user.id
    tg_username = message.from_user.username

    if not tg_username:
        await message.answer(
            "⚠️ У вас не установлен username в Telegram.\n"
            "Пожалуйста, установите его в настройках Telegram и повторите /start."
        )
        return

    username = normalize_username(tg_username)
    contact_phone = normalize_phone(message.contact.phone_number)

    profile = api_get_user_profile(username)
    if not profile:
        await message.answer("🚫 Пользователь не найден на сайте. Напишите в поддержку.")
        await offer_support(message)
        return

    site_phone = normalize_phone(profile.get("phone_number") or "")
    if not site_phone:
        await message.answer("⚠️ В профиле на сайте не указан телефон. Напишите в поддержку.")
        await offer_support(message)
        return

    if contact_phone != site_phone:
        await message.answer("🚫 Телефон не совпадает с профилем сайта. Напишите в поддержку.")
        await offer_support(message)
        return

    user_id = profile.get("id")
    if isinstance(user_id, int):
        api_patch_user(user_id, {"telegram_id": tg_user_id})

    await message.answer("✅ Доступ подтверждён.")
    await ask_enable_notifications(message)

async def ask_enable_notifications(message: types.Message):
    kb = InlineKeyboardMarkup()
    kb.add(InlineKeyboardButton(text='Да ✅', callback_data='notify_yes'))
    kb.add(InlineKeyboardButton(text='Нет ❌', callback_data='notify_no'))
    await message.answer('🔔 Хотите получать уведомления от сайта?', reply_markup=kb)

@dp.callback_query_handler(lambda c: c.data and c.data.startswith('notify_'))
async def on_notify_choice(callback_query: types.CallbackQuery):
    tg_user_id = callback_query.from_user.id
    if callback_query.data == 'notify_yes':
        set_notify_enabled(tg_user_id, True)
        await callback_query.answer('Уведомления включены!')
        await bot.send_message(tg_user_id, "🔔 Уведомления включены. (Пока тестовый режим)")
    else:
        set_notify_enabled(tg_user_id, False)
        await callback_query.answer('Уведомления выключены.')

# -----------------------------
# Поддержка (пересылка админу)
# -----------------------------

async def offer_support(message: types.Message):
    await message.answer("✉️ Опишите проблему одним сообщением — оно будет переслано в поддержку.")

@dp.message_handler(lambda m: m.text and not m.text.startswith('/'))
async def forward_to_support(message: types.Message):
    if ADMIN_CHAT_ID and ADMIN_CHAT_ID != 0:
        await bot.send_message(
            ADMIN_CHAT_ID,
            f"Сообщение в поддержку от {message.from_user.full_name} (@{message.from_user.username}, id={message.from_user.id}):\n{message.text}"
        )
        await message.answer("✅ Сообщение отправлено в поддержку.")
    else:
        await message.answer("⚠️ Поддержка не настроена (ADMIN_CHAT_ID).")

# -----------------------------
# Заглушка уведомлений: периодический опрос /notification/all
# -----------------------------

async def notifications_poller():
    """
    Раз в 15 секунд читаем уведомления, отправляем новые всем пользователям, кто включил notify.
    Это MVP. На бэке нет поля 'sent' и адресности — поэтому храним last_id локально.
    """
    import asyncio

    while True:
        try:
            last_id_str = get_bot_state("last_notification_id") or "0"
            last_id = int(last_id_str)

            data = api_get_notifications(limit=50, offset=0)
            if data and "items" in data:
                items = data["items"]
                new_items = [x for x in items if isinstance(x, dict) and x.get("id", 0) > last_id]
                new_items.sort(key=lambda x: x.get("id", 0))

                if new_items:
                    conn = db_conn()
                    cur = conn.cursor()
                    cur.execute("SELECT tg_user_id FROM user_state WHERE notify_enabled=1")
                    subs = [row["tg_user_id"] for row in cur.fetchall()]
                    conn.close()

                    for n in new_items:
                        text = n.get("text") or n.get("title") or str(n)
                        for uid in subs:
                            try:
                                await bot.send_message(uid, f"🔔 Уведомление: {text}")
                            except Exception:
                                logging.exception("Failed to send notification to %s", uid)

                    set_bot_state("last_notification_id", str(new_items[-1].get("id", last_id)))

        except Exception:
            logging.exception("notifications_poller loop error")

        await asyncio.sleep(15)

async def on_startup(_):
    import asyncio
    asyncio.create_task(notifications_poller())

if __name__ == '__main__':
    import asyncio

    loop = asyncio.new_event_loop()
    asyncio.set_event_loop(loop)

    executor.start_polling(dp, skip_updates=True, on_startup=on_startup)