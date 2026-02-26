server:
  port: 8080
  mode: release
  php_backend: "http://127.0.0.1:9000"
  php_public_url: ""
  bridge_secret: "{{bridge_secret}}"

database:
  host: 127.0.0.1
  port: 3306
  user: {{db_user}}
  password: "{{db_pass}}"
  dbname: "{{db_name}}"
  max_open_conns: 50
  max_idle_conns: 25

redis:
  host: 127.0.0.1
  port: 6379
  password: "{{redis_pass}}"
  db: 0

jwt:
  secret: "{{jwt_secret}}"
  access_ttl: 604800
  refresh_ttl: 2592000

cache:
  order_list_ttl: 30
  class_list_ttl: 300

smtp:
  host: "smtp.qq.com"
  port: 465
  user: ""
  password: ""
  from_name: "系统通知"
  encryption: "ssl"

license:
  license_key: ""
  domain: ""
  key_file: ""
  cache_file: ".sys_state"
  secrets_file: ".secrets"
