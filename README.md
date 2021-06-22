Сервис перенаправления и доступа к http сервисам 1С из мира
---

Данный сервис "перенаправляет" запросы в 1С. Со всеми параметрами и по всем адресам.
Сервис работает по адресу http://127.0.0.1:8087.
Перенаправляется в мир через NGINX на 80 порт (полный конфиг по адресу /etc/nginx/sites-available/access-vmpauto-space.conf).

---

Запускается сервисом

`service access-vmpauto-space start`

Перезапускается сервисом

`service access-vmpauto-space restart`

Конфигурирование сервиса в катологе

`/etc/systemd/system/access-vmpauto-space.service`

---

После правок необходимо убить сервис и его процессы

`service access-vmpauto-space stop`

`kill -9 [id процесса]`
 
и собрать заново `go build`

`service access-vmpauto-space start`

---

**ПОЛЬЗОВАТЕЛИ ДОЛЖНЫ БЫТЬ СВЯЗАНЫ С ДОМЕНОМ!!!**

Точка входа устанавливается в константу `DefaultPoint1C` в api.go.
В глобальной переменной `views` устанавливаются параметры сервиса, который собирает
BasicAuth из 1С по доменному логину.

`
var views = newHTTPService(DefaultPoint1C, "http_service", "бла-бла-бла")
`

Ручка в 1С `/service/token?domainuser=`, где domainuser - логин в домене.
Ручка в 1С для токена с собственным паролем`/service/token?domainuser=user&password=PaSsWoRd`, где **domainuser** - логин в домене, **password** - действительный пароль. 
Получившийся токен добавляем в cache.json
 
Код ручки в 1С в справочнике WebСтраницы:
```1CEnterprice

    // - точка входа - /service/token
    // - заголовки   - application/json

    // - имя в домене
    ПользовательДомена = Запрос.ПараметрыЗапроса.Получить("domainuser");
    // - если пароль отличается от пароля по умолчанию
    ПарольПользователя = Запрос.ПараметрыЗапроса.Получить("password");
    // - пароль по умолчанию
    ПарольДляВсехПоУмолчанию = ?(ПарольПользователя=Неопределено, "135246", ПарольПользователя);

    Соответствие = Новый Соответствие();

    // - перебираем активных пользователей
    Для каждого мПользователь Из ПользователиИнформационнойБазы.ПолучитьПользователей() Цикл
        
        // - если не нашли
        Если ПользовательДомена <> Нрег(СтрЗаменить(Строка(мПользователь.ПользовательОс),"\\VMPAUTO\","")) Тогда
            Продолжить;
        КонецЕсли;
        
        // - пишем в память, дабы не ваять кучу файлов
        ПотокВПамяти = Новый ПотокВПамяти();
        ЗаписьДанных = Новый ЗаписьДанных(ПотокВПамяти);
        ЗаписьДанных.ЗаписатьСтроку(Строка(мПользователь.Имя)+":"+ПарольДляВсехПоУмолчанию);
        ДвоичныеДанные = ПотокВпамяти.ЗакрытьИПолучитьДвоичныеДанные();

        // - пишем
        Соответствие[Нрег(СтрЗаменить(Строка(мПользователь.ПользовательОс),"\\VMPAUTO\",""))] = Base64Строка(ДвоичныеДанные);
        
    КонецЦикла; 

    // - кодируем и отправляем
    ПараметрыЗаполненияСтраницы.Вставить("СтрокаРезультата", МониторОбмена.ЗаписатьСтрокуJSON(Соответствие));

```

---

Структура проекта:
- /static/ - вся статика
    - /static/icons/ - каталог с иконками        
    - /static/2fa.html - настройка 2-фа по qr коду
    - /static/first.html - авторизация по домену и начало регистрации
    - /static/index.html - стартовая со входом по логину и одоразовому паролю
    - /static/manual2fa.html - ручная регистрация 2-фа по ссылке или данным
    - /static/message.html - отображение простеньких сообщений

- api.go - методы по взаимодействию с 1С
- cache.go - кэш приложения и работа с ним
- cache.json - бэкап кэша с ключами
- helpers.go - вспомогательные методы
- logger.go - логирование
- main.go - точка входа в приложение и ручки на вход
- reg.go - ручки иметоды регистрации
- sessions.go - работа с токенами, куками
- tokens.json - бэкап токенов и кук

