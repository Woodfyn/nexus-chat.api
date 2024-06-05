# CHAT-API 

## Description.

This is a common messenger that has private and group chats. You can also edit your chat profile after registration. The database is PostgreSQL. I used AWS S3 for file storage and deployed the application on AWS EC2. I also used Twilio to verify phone numbers and web sockets for chat

## How to start.
Build debug. Write in console:
    ```
    make build-debug && run
    ```
Build prod. Write in console:
    ```
    make build-prod && run
    ```
Build all. Write in console:
    ```
    make build && run
    ```
