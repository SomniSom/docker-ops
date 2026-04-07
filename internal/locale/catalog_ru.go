package locale

var catalogRu = map[string]string{
	"root.short": "Docker Compose локально и по SSH",
	"root.long":  "%s (%s) — управление проектами Docker Compose. Полная спецификация — в readme.ru.md.",

	"flag.lang":         "Язык интерфейса: en, ru или auto (по LANG / LC_MESSAGES / DQ_LANG)",
	"flag.project_root": "корень проекта (по умолчанию: $DQ_PROJECT_ROOT или текущий каталог)",

	"version.short": "Показать сведения о версии",

	"validate.short": "Проверить синтаксис и поля docker-ops.yaml / docker-ops.yml",
	"validate.long": `Разбирает конфиг в корне проекта и выдаёт понятные ошибки при неверном YAML
(частая ошибка: пробел перед корневым ключом вроде deploy_include). Также проверяет deploy_mode
и наличие app_config на диске, если он задан.`,
	"validate.err.no_docker_ops": "в %s нет docker-ops.yaml и docker-ops.yml",
	"validate.note.remote":       "\ndq: замечание: заданы remote_ssh и remote_path — команды compose выполняются по SSH. Только локальный Docker: DOCKER_OPS_USE_REMOTE=0 или use_remote: false в YAML.\n",

	"configcheck.short":        "Проверить файл app_config, если он задан (readme §5.3)",
	"configcheck.info.not_set": "dq: app_config не задан — проверять нечего",
	"msg.ok_path":              "ок: %s",

	"deploy.short": "Деплой проекта на удалённый хост по SSH (readme §5.3)",
	"deploy.long": `Доставка файлов на сервер по SFTP (режим source) или compose + образ (artifacts), затем удалённый docker compose (readme §5.3).

Source (по умолчанию): зеркалирование дерева с исключениями (размер/mtime), deploy_include, app_config, затем удалённый reup.

Artifacts: нужны docker-compose.image.yml и deploy_image; опционально docker build/push или save|ssh|load; затем удалённый pull+up или up. Файл compose: dq gen-image-compose. Флаг --build выполняет docker build -t deploy_image перед save/load (или push при registry), даже если deploy_push не задан.`,
	"deploy.err.needs_remote": "для deploy нужны remote_ssh и remote_path (docker-ops.yml или dq.env); выполните «dq validate»",
	"deploy.flag.build":       "artifacts: docker build -t deploy_image, затем save/load или push по настройкам",

	"build.short":   "docker compose build --pull",
	"pull.short":    "docker compose pull",
	"up.short":      "docker compose up -d",
	"down.short":    "docker compose down",
	"stop.short":    "docker compose stop (контейнеры остаются; опционально имена сервисов)",
	"reup.short":    "docker compose build --pull && up -d",
	"ps.short":      "docker compose ps",
	"restart.short": "docker compose restart <сервис>",
	"restart.err":   "compose_service пуст и сервис не указан в аргументах",

	"exec.short": "docker compose exec (интерактивный TTY, если stdin — терминал)",
	"exec.long": `Команда внутри контейнера сервиса. Если stdin — TTY, вызывается docker compose exec -it (например dq exec audio-bot bash).

Ctrl+C передаётся в контейнер (прервать текущую строку). Ctrl+D (EOF) — выход из shell. Без интерактивного терминала используется exec -T. Чтобы отключить псевдо-TTY в терминале, укажите -T / --no-tty.`,
	"exec.flag.no_tty": "docker compose exec -T (без псевдо-TTY)",

	"status.short":       "docker compose ps -a и последние логи (по умолчанию все сервисы)",
	"status.header.ps":   "=== docker compose ps -a (проект %s) ===\n",
	"status.header.logs": "\n=== последние 80 строк логов (%s) ===\n",
	"status.err.logs":    "dq: логи: %v\n",
	"status.label.all":   "все сервисы",

	"logs.use.follow":   "logs [сервис...]",
	"logs.use.tail":       "logs-tail [сервис...]",
	"logs.short.follow":   "docker compose logs -f (все сервисы, если имена не указаны)",
	"logs.short.tail":     "docker compose logs --tail 200 (все сервисы, если имена не указаны)",
	"logs.long": `Поток или вывод логов compose. Без имён сервисов — логи всех сервисов.

Если stdin не терминал (например запуск из IDE по SSH), вместо -f используется --tail 200, чтобы логи не были пустыми.

Имена сервисов — позиционные аргументы перед флагами, например:
  dq logs
  dq logs parser
  dq logs parser worker --tail 50`,

	"env.short": "Вывести шаблон docker-ops.yaml (readme §5.4)",
	"env.err":   "отказ перезаписать %s (используйте --force)",
	"env.wrote": "dq: записано %s\n",

	"completion.short": "Сгенерировать скрипт автодополнения оболочки",
	"completion.long": `Подключение автодополнений:

Bash:
  source <(dq completion bash)

Zsh:
  source <(dq completion zsh)

Fish:
  dq completion fish | source
`,

	"paths.err.app_config": "конфиг приложения %s",
	"paths.err.app_dir":    "конфиг приложения — это каталог: %s",

	"err.remote_not_configured": "удалённый доступ не настроен",
	"err.remote_ssh":            "удалённый SSH не настроен",
	"err.config_required":       "нужна конфигурация",

	"load.err.read":   "чтение конфига",
	"load.err.config": "конфиг %s",

	"validate.err.read_file": "чтение %s",
	"validate.deploy_mode":        "deploy_mode должен быть «source», «artifacts» или пустым; получено %q",
	"validate.app_config_missing": "app_config указывает на отсутствующий файл %q (полный путь: %s)",
	"validate.err.dq_env": "dq.env",

	"envfile.expected_kv": "%s:%d: ожидалось KEY=value",
	"envfile.empty_key":   "%s:%d: пустой ключ",

	"yaml.invalid_intro":      "некорректный YAML docker-ops в %s\n\n",
	"yaml.parser_said":        "Сообщение парсера: %s\n",
	"yaml.context_header":     "\nКонтекст (строка парсера %d):\n",
	"yaml.past_eof":           "\n(Указанная строка %d за концом файла; показан хвост.)\n",
	"yaml.no_line":            "\nНе удалось сопоставить ошибку со строкой; проверьте отступы и двоеточия в ключах.\n",
	"yaml.common_issues":      "Типичные причины:\n",
	"yaml.issue.root_keys":    "  • Корневые ключи (remote_ssh, deploy_include, …) должны начинаться с колонки 1 — без ведущих пробелов.\n",
	"yaml.issue.spaces_lists": "  • В списках (например deploy_include) используйте пробелы для отступа, не смешивайте табы и пробелы.\n",
	"yaml.issue.colon_quote":  "  • Значения с ':' лучше заключать в кавычки.\n",
	"yaml.misindented_root":   "Обнаружен с отступом корневой ключ %q в строке %d — корневые ключи должны начинаться с колонки 0 (без пробелов и табов).",
	"yaml.hint.tabs":          "Подсказка: в строке %d перед %q используются табы. Лучше только пробелы или исправьте отступ.",
	"yaml.hint.spaces":        "Подсказка: %q сдвинут на %d пробел(а/ов). Корневые ключи в docker-ops.yml должны быть в начале строки (колонка 0). Уберите пробелы перед %q.",
	"yaml.hint.key_space":     "Подсказка: в ключе %q может быть лишний пробел; в YAML нельзя неэкранированные пробелы в ключе перед ':'.",
	"yaml.hint.top_level":     "Корневые ключи",

	"compose.docker_path": "docker не найден в PATH",
	"compose.v2_required": `Нужен плагин Docker Compose V2 (команда: docker compose).
Отдельный docker-compose (V1) не поддерживается.

%s

Сообщение Docker: %s`,
	"compose.version_prefix": "docker compose version",
	"compose.install_hint": `Установите плагин Compose V2, например:
  • Документация: https://docs.docker.com/compose/install/linux/
  • Debian/Ubuntu: sudo apt-get update && sudo apt-get install -y docker-compose-plugin
  • Проверка: docker compose version`,
	"compose.run_prefix": "docker compose",
	"compose.file_missing": "файл compose %s",

	"template.header1": "# docker-ops — Docker Quick-ops (сгенерировано dq env)",
	"template.header2": "# Сохраните как docker-ops.yaml в корне проекта (секреты — в dq.env, не коммитьте)",

	"deploy.src.remote_mkdir": "удалённый mkdir",
	"deploy.src.sftp":         "sftp",

	"deploy.art.docker_build": "docker build",
	"deploy.art.docker_push":  "docker push",
	"deploy.art.err.image":    "deploy_mode=artifacts требует deploy_image (docker-ops.yml или dq.env)",
	"deploy.art.err.compose":  "деплой artifacts требует %s в корне проекта",
	"deploy.art.upload":       "загрузка compose",
	"deploy.art.remote_up_sl": "==> удалённо: compose up (образ загружен локально)\n",
	"deploy.art.remote_pull":  "==> удалённо: compose pull && up\n",
	"deploy.art.save_gzip":    "==> docker save (gzip) | ssh … docker load\n",
	"deploy.art.save_plain":   "==> docker save | ssh … docker load\n",
	"deploy.art.build":        "==> docker build -t %s\n",
	"deploy.art.push":         "==> docker push %s\n",

	"deploy.inc.skip_abs":    "dq: deploy_include: пропуск абсолютного пути %q\n",
	"deploy.inc.skip_unsafe": "dq: deploy_include: пропуск небезопасного пути %q\n",
	"deploy.inc.missing":     "dq: предупреждение: deploy_include: нет %s\n",
	"deploy.inc.err": "deploy_include %s",

	"deploy.mirror.list":   "список удалённых файлов",
	"deploy.mirror.remove": "удаление на сервере %s",
	"deploy.mirror.mkdir":  "mkdir %s",
	"deploy.mirror.upload": "загрузка %s",
	"deploy.mirror.not_under": "%q не внутри %q",

	"deploy.appcfg.warn": "dq: предупреждение: локально нет app_config (%s); копирование на сервер пропущено\n",
	"deploy.appcfg.dir":  "app_config — это каталог: %s",

	"ssh.err.no_credentials": "нет учётных данных SSH: укажите ssh_identity на незашифрованный ключ или используйте ssh-agent (зашифрованные ключи)",
	"ssh.err.empty_remote":   "пустой remote_ssh",
	"ssh.err.bad_remote":     "remote_ssh должен быть user@host (получено %q)",
	"ssh.err.no_host":        "в remote_ssh после @ нет хоста",
	"ssh.err.known_hosts":    "known_hosts",
	"ssh.err.kh_file":        "known_hosts %s",
	"ssh.err.host_key_suffix": "отказ: ключ хоста изменился или не совпадает — проверьте MITM",
	"ssh.err.dial":           "ssh dial %s",
	"ssh.err.remote":         "удалённо",
	"ssh.err.stdin":          "нужны ssh-клиент и stdin",
	"ssh.err.request_pty":    "запрос pty",

	"genimg.short": "Собрать docker-compose.image.yml из compose_file для деплоя artifacts",
	"genimg.long": `Переводит сервис приложения с build: на image: (по умолчанию compose_service). Сервисы только с image: не меняются.

Ошибка, если у другого сервиса остаётся build: (флаг --all-built — только если несколько сборок с одним тегом).

Примеры:
  dq gen-image-compose
  dq gen-image-compose --service app -o docker-compose.image.yml`,
	"genimg.header": "# Сгенерировано dq gen-image-compose — задайте DEPLOY_IMAGE при деплое (docker-ops.yml / dq.env).",
	"genimg.wrote":  "записано %s",
	"genimg.err.read":       "чтение %s",
	"genimg.err.no_service": "укажите compose_service в docker-ops или передайте --service",
	"genimg.err.transform":  "преобразование compose",
	"genimg.err.write":      "запись %s",
	"genimg.flag.output":        "путь вывода относительно корня проекта",
	"genimg.flag.compose_file":  "входной compose относительно корня (по умолчанию compose_file из конфига)",
	"genimg.flag.service":       "сервис для конвертации (по умолчанию compose_service)",
	"genimg.flag.image_expr":    "значение для image: (по умолчанию ${DEPLOY_IMAGE})",
	"genimg.flag.all_built":     "конвертировать все сервисы с build: в один и тот же образ",

	"man.short": "Открыть man-страницу (groff) для dq или подкоманды",
	"man.long": `Страница собирается из справки Cobra и показывается через man -l (нужны man-db / groff).

Примеры:
  dq man
  dq man deploy
  dq man gen-image-compose

Установка в систему: make install-man (см. Makefile).`,
	"man.err.gen":     "генерация man-страницы",
	"man.hint.no_man": "\n(man не найден в PATH; выше — исходник troff — можно: groff -man -T utf8 | less)",
}
