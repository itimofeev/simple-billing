# Тестовое задание Billing: Микросервис баланса пользователей

## Задание
Приложение хранит в себе идентификаторы пользователей и их баланс. Взаимодействие с ним осуществляется исключительно с помощью брокера очередей.

По требованию внешней системы, микросервис может выполнить одну из следующих операций со счетом пользователя:
- Списание
- Зачисление
- Перевод от пользователя к пользователю
- (будет плюсом, но не обязательно) Блокирование с последующим списанием или разблокированием. Заблокированные средства недоступны для использования. Блокировка означает что некая операция находится на авторизации и ждет какого-то внешнего подтверждения, ее можно впоследствии подтвердить или отклонить

После проведения любой из этих операций генерируется событие-ответ в одну из очередей.

Основные требования к воркерам:
- Код воркеров должен безопасно выполняться параллельно в разных процессах
- Воркеры могут запускаться одновременно в любом числе экземпляров и выполняться произвольное время
- Все операции должны обрабатываться корректно, без двойных списаний, отрицательный баланс не допускается

В пояснительной записке к выполненному заданию необходимо указать перечень используемых инструментов и технологий, способ развертки приложения, общий механизм работы (интерфейсы ввода/вывода)

Будет плюсом покрытие кода юнит-тестами.

Требования к окружению:
- Язык программирования: PHP 7 (PSR-2) либо Go 1.10+
- Можно использовать: любые фреймворки, реляционные БД для хранения баланса, брокеры очередей, key-value хранилища.

