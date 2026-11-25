# 🎨 CampBooking Frontend

Современный фронтенд для системы бронирования детских лагерей.

---

## 📦 Состав

### 1. **index.html** - Landing Page
Главная страница сайта с:
- Hero section с анимациями
- Секция преимуществ (6 карточек)
- Каталог популярных лагерей
- Footer
- Адаптивный дизайн

**Стек:** React 18, Vanilla CSS, Bebas Neue + Manrope fonts

---

### 2. **auth.html** - Авторизация
Страница входа/регистрации с:
- Переключение между вкладками Login/Register
- Интеграция с API (`POST /api/v1/login`, `POST /api/v1/register`)
- JWT токены в localStorage
- Валидация форм
- Анимированный сайдбар

---

### 3. **dashboard.html** - Личный кабинет
Две роли с разным интерфейсом:

**Для администратора:**
- Статистика (лагеря, смены, бронирования, выручка)
- Управление лагерями (CRUD)
- Модальное окно создания лагеря
- Таблица с данными

**Для родителя:**
- Статистика (бронирования, дети, дни до начала)
- Мои бронирования
- Скачивание PDF путёвок
- Кнопка "Забронировать"

---

## 🎨 Дизайн

### Цветовая палитра
```css
--primary: #FF6B35    /* Оранжевый (CTA, кнопки) */
--secondary: #004E89  /* Синий (заголовки, текст) */
--accent: #FFD23F     /* Жёлтый (акценты) */
--success: #06D6A0    /* Зелёный (успех) */
--error: #EF476F      /* Красный (ошибки) */
```

### Типографика
- **Display:** Bebas Neue (заголовки)
- **Body:** Manrope (основной текст)

### Особенности дизайна
✅ Bold, уникальный стиль (не generic AI)  
✅ Анимации (fade-in, slide-up, pulse)  
✅ Градиенты и тени  
✅ Адаптивность (mobile-first)  
✅ Hover эффекты  
✅ Модальные окна  

---

## 🚀 Запуск

### Вариант 1: Просто откройте файлы
```bash
# Откройте index.html в браузере
open index.html
```

### Вариант 2: Локальный сервер
```bash
# Python
python3 -m http.server 8000

# Node.js
npx http-server

# Затем откройте http://localhost:8000
```

### Вариант 3: Интеграция с бэкендом
Поместите HTML файлы в Nginx:

```bash
# nginx.conf
server {
    listen 80;
    root /usr/share/nginx/html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://gateway:8080;
    }
}
```

Структура:
```
/usr/share/nginx/html/
├── index.html        ← Landing
├── auth.html         ← Авторизация
└── dashboard.html    ← Кабинет
```

---

## 🔌 API Интеграция

### Конфигурация
В каждом файле есть:
```javascript
const API_BASE = 'http://localhost/api/v1';
```

Измените на свой URL при деплое.

### Endpoints используемые
```
POST /api/v1/register   - Регистрация
POST /api/v1/login      - Авторизация
GET  /api/v1/camps      - Список лагерей
POST /api/v1/camps      - Создать лагерь (Admin)
```

### JWT Токены
Токен сохраняется в `localStorage`:
```javascript
localStorage.setItem('token', data.token);

// Использование
fetch('/api/v1/camps', {
    headers: {
        'Authorization': `Bearer ${token}`
    }
});
```

---

## 📱 Responsive Design

Breakpoints:
- **Desktop:** > 968px (полный функционал)
- **Tablet:** 768px - 968px (адаптированная сетка)
- **Mobile:** < 768px (вертикальные карточки, скрытый sidebar)

---

## ✨ Фичи

### Landing Page
- ✅ Анимированный Hero
- ✅ Пульсирующий фон
- ✅ Hover эффекты на карточках
- ✅ Smooth scroll

### Auth Page
- ✅ Переключение Login/Register
- ✅ Real-time валидация
- ✅ Alerts (success/error)
- ✅ Loading states

### Dashboard
- ✅ Фиксированный sidebar
- ✅ Статистика в реальном времени
- ✅ Модальные окна (создание лагеря)
- ✅ Таблицы с hover
- ✅ Badge статусов
- ✅ Logout функция

---

## 🎯 TODO (Будущие улучшения)

- [ ] Страница каталога всех лагерей с фильтрами
- [ ] Процесс бронирования (multi-step form)
- [ ] Профиль пользователя
- [ ] Календарь доступных смен
- [ ] Чат поддержки
- [ ] Push уведомления
- [ ] Тёмная тема
- [ ] i18n (мультиязычность)

---

## 🐛 Известные проблемы

- Нет обработки 401 ошибок (logout при истечении токена)
- Нет offline режима
- Модальное окно не закрывается по ESC

---

## 💡 Советы по кастомизации

### Сменить цвета
Измените CSS переменные в `:root`:
```css
:root {
    --primary: #YOUR_COLOR;
    --secondary: #YOUR_COLOR;
}
```

### Добавить новую страницу
1. Скопируйте структуру любого HTML
2. Добавьте свой контент в `<script type="text/babel">`
3. Используйте существующие CSS классы

### Добавить endpoint
```javascript
const myFunction = async () => {
    const token = localStorage.getItem('token');
    const response = await fetch(`${API_BASE}/your-endpoint`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify(data)
    });
    const result = await response.json();
};
```

---

## 📸 Скриншоты

### Landing Page
- Яркий градиентный Hero
- Карточки преимуществ с иконками
- Каталог лагерей

### Dashboard
- Sidebar с навигацией
- Статистика в карточках
- Таблица данных

---

## 🤝 Контакты

**Проект:** CampBooking  
**Репозиторий:** https://gitlab.crja72.ru/golang/2026/spring/projects/go10/camp-booking  

---

**Готово к использованию!** 🚀
