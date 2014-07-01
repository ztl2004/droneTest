update
======

The update module of Ark
How to build?
=============

The project build with [redis](https://github.com/hoisie/redis),[martini](https://github.com/go-martini/martini), [martini-render](https://github.com/martini-contrib/render), 
[mysql](https://github.com/go-sql-driver/mysql) and
[xorm](https://github.com/go-xorm/xorm),

Install [redis](https://github.com/hoisie/redis)
,[martini](https://github.com/go-martini/martini)
,[martini-render](https://github.com/martini-contrib/render)
,[mysql](https://github.com/go-sql-driver/mysql)
and [xorm](https://github.com/go-xorm/xorm)
```
go get -u github.com/hoisie/redis
go get -u github.com/go-martini/martini
go get -u github.com/martini-contrib/render
go get -u github.com/go-sql-driver/mysql
go get -u github.com/go-xorm/xorm
```
How to Initlization MySQL Database?
===================================

```
INSERT INTO mysql.user(Host,User,Password) VALUES ('localhost', 'arkors', password('arkors'));
FLUSH PRIVILEGES;
CREATE DATABASE `arkors_update` DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;
GRANT ALL PRIVILEGES ON arkors_update.* TO arkors@localhost IDENTIFIED BY 'arkors';
FLUSH PRIVILEGES;
```

Before run go test
===================================
you need to set your mysql root/password and redis's port in update_test.go

```
root_db, err := sql.Open("mysql", "root:@/")

```

