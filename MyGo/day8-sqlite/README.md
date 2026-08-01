# День 8: SQLite — постоянное хранение данных

В Дне 7 мы хранили задачи в оперативной памяти (глобальный слайс). При перезапуске сервера все данные терялись.  
Сегодня мы подключили **SQLite** — лёгкую встраиваемую базу данных, которая хранит данные в файле `tasks.db`. Теперь задачи сохраняются даже после выключения сервера.

# Что сделано

- Установлен драйвер SQLite: `github.com/mattn/go-sqlite3`.
- Создана таблица `tasks` с полями: `id`, `title`, `description`, `done`.
- Переписаны все обработчики (`getTasksHandler`, `createTaskHandler` и др.) на работу с базой данных.
- Удалены глобальные переменные `tasks` и `idCont` — теперь данные живут в БД.

# Структура базы данных

Таблица `tasks` создаётся автоматически при первом запуске:

```sql
CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    done BOOLEAN DEFAULT 0
);