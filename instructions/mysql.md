For MySQL service in the template, use this definition:

```yaml
- name: MySQL
  icon: https://raw.githubusercontent.com/zeabur/service-icons/main/marketplace/mysql.svg
  template: PREBUILT
  spec:
    source:
        image: mysql:8.0.33
    ports:
        - id: database
          port: 3306
          type: TCP
    volumes:
        - id: data
          dir: /var/lib/mysql
    instructions:
        - type: TEXT
          title: Command to connect to your MySQL
          content: mysqlsh --sql --host=${PORT_FORWARDED_HOSTNAME} --port=${DATABASE_PORT_FORWARDED_PORT} --user=${MYSQL_USERNAME} --password=${MYSQL_PASSWORD} --schema=${MYSQL_DATABASE}
        - type: TEXT
          title: MySQL username
          content: ${MYSQL_USERNAME}
          category: Credentials
        - type: PASSWORD
          title: MySQL password
          content: ${MYSQL_PASSWORD}
          category: Credentials
        - type: TEXT
          title: MySQL database
          content: ${MYSQL_DATABASE}
          category: Credentials
        - type: TEXT
          title: MySQL host
          content: ${PORT_FORWARDED_HOSTNAME}
          category: Hostname & Port
        - type: TEXT
          title: MySQL port
          content: ${DATABASE_PORT_FORWARDED_PORT}
          category: Hostname & Port
    env:
        MYSQL_DATABASE:
            default: zeabur
            expose: true
        MYSQL_HOST:
            default: ${CONTAINER_HOSTNAME}
            expose: true
            readonly: true
        MYSQL_PASSWORD:
            default: ${MYSQL_ROOT_PASSWORD}
            expose: true
            readonly: true
        MYSQL_PORT:
            default: ${DATABASE_PORT}
            expose: true
            readonly: true
        MYSQL_ROOT_PASSWORD:
            default: ${PASSWORD}
        MYSQL_USERNAME:
            default: root
            expose: true
            readonly: true
    configs:
        - path: /etc/my.cnf
          template: |
            [mysqld]
            default-authentication-plugin=mysql_native_password
            skip-host-cache
            skip-name-resolve
            datadir=/var/lib/mysql
            socket=/var/run/mysqld/mysqld.sock
            secure-file-priv=/var/lib/mysql-files
            user=mysql
            max_allowed_packet=10M
            performance_schema=off

            pid-file=/var/run/mysqld/mysqld.pid
            
            [client]
            socket=/var/run/mysqld/mysqld.sock

            !includedir /etc/mysql/conf.d/
```

For other services requiring MySQL connectivity, utilize the exposed environment variables from MySQL's env section.
