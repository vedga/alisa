# alisa
Поддержка сервиса Алиса

Описание протокола см. https://yandex.ru/dev/dialogs/smart-home/doc/reference/resources.html

Публикация навыка: https://dialogs.yandex.ru/developer

Консоль для регистрации OAuth-приложений (не требуется): https://oauth.yandex.ru/

Проверка сертификата и пути сертификации:

openssl s_client -connect iot.domain.com:8443

Если использовать авторизацию через Yandex oAuth, то для IoT в качестве callback URL необходимо указывать https://social.yandex.net/broker/redirect, в связке аккаунтов в поле "URL авторизации" указывать https://oauth.yandex.ru/authorize, в связке аккаунтов в поле "URL для получения токена" указывать https://oauth.yandex.ru/token (идентификатор клиента и секретный ключ берется со страницы, на которой регистрировали oAuth в Yandex). В этом случае не придется реализовывать oAuth самостоятельно.
