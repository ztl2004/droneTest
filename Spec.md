## Update API Specification


### API Specification

#### Update API Specification

##### POST /v1/updates/:app

创建一个 Application 的版本信息。由 [Board](https://github.com/arkors/board) 模块通过 UI 界面进行调用。

###### Example Request:
```
POST /v1/updates/232 HTTP/1.1
Host: update.arkors.com
X-Arkors-Application-Log: 5024442115e7bd738354c1fac662aed5
X-Arkors-Application-Client: xxx.xx.xx.xxx,BOARD
Accept: application/json
{
  "version": 3,
  "name": "0.3.3",
  "update": "2014-11-10T00:00:00.000Z",
  "changed": "1. New design application icon.\n 2. Fix some bugs.",
  "url": "http://file.arkors.com/releases/demo-lastest.apk",
  "client": "Android",
  "compatible": "0.3.2 0.3.1 0.2.x 0.1.x"    
}
```

###### Example Response
```
HTTP/1.1 201 OK
X-Arkors-Application-Log: cb21df532c6647383af7efa0fd8405f2
Content-Type: application/json
{
  "app": 232,
  "version": 3,
  "name": "0.3.3",
  "update": "2014-11-10T00:00:00.000Z",
  "changed": "1. New design application icon.\n 2. Fix some bugs.",
  "url": "http://file.arkors.com/releases/demo-lastest.apk",
  "client": "Android",
  "compatible": "0.3.2 0.3.1 0.2.x 0.1.x"    
}
```

###### Status Codes:
* 201 – 创建 Update 记录成功
* 400 – Errors (invalid json, missing, duplication or invalid fields, etc)

##### GET /v1/updates/:app/:version

客户端访问 /v1/updates/:app ，发送当前运行版本号的ID，由服务器判断是否有最新版本。如果有最新版本，返回版本的相关信息，由 Mobile 的逻辑判断是否升级；如果没有新版本返回 HTTP 的 404 状态码。

判断规则：

* 判断从 HTTP HEADER 中获取的客户端类型和版本信息中的客户端类型是否一致。
* 判断当前版本是否比最新版本低。
* 判断当前版本是否在可升级版本中。
* 判断当前时间是否在可升级时间内。

###### Example Request:
```
GET /v1/updates/232/3 HTTP/1.1
Host: update.arkors.com
X-Arkors-Application-Id: 232
X-Arkors-Application-Token: cb21df532c6647383af7efa0fd8405f2,1389085779854
X-Arkors-Application-Log: 5024442115e7bd738354c1fac662aed5
X-Arkors-Application-Client: 3ad3ce877d6c42b131580748603f8d6a,ANDROID
Accept: application/json
```

###### Example Response
```
HTTP/1.1 200 OK
X-Arkors-Application-Log: cb21df532c6647383af7efa0fd8405f2
Content-Type: application/json
{
  "app": 323,
  "version": 3,
  "name": "0.3.3",
  "update": "2014-11-00T00:00:00.000Z",
  "changed": "1. New design application icon.\n 2. Fix some bugs.",
  "url": "http://file.arkors.com/releases/demo-lastest.apk",
  "client": "Android",
  "compatible": "0.3.2 0.3.1 0.2.x 0.1.x"    
}
```

###### Status Codes:
* 200 - 返回 Update 登录信息。
* 400 - Errors (invalid json, missing or invalid fields, etc)
* 401 - Unauthorized，把 Token 发送到 OAuth 模块认证返回失败信息。
* 404 - 没有要升级的版本信息

##### PUT /v1/updates/:app/:version

更新一个 Application 的版本信息。由 [Board](https://github.com/arkors/board) 模块通过 UI 界面进行调用。

###### Example Request:
```
PUT /v1/updates/232/3 HTTP/1.1
Host: update.arkors.com
X-Arkors-Application-Log: 5024442115e7bd738354c1fac662aed5
X-Arkors-Application-Client: xxx.xx.xx.xxx,BOARD
Accept: application/json
{
  "version": 3,
  "name": "0.3.3",
  "update": "2014-11-02T00:00:00.000Z",
  "changed": "1. New design application icon.\n 2. Fix some bugs. 3. Add GCM push supported",
  "url": "http://file.arkors.com/releases/demo-lastest.apk",
  "client": "Android",
  "compatible": "0.3.2 0.3.1 0.2.x 0.1.x"    
}
```

###### Example Response
```
HTTP/1.1 200 OK
X-Arkors-Application-Log: cb21df532c6647383af7efa0fd8405f2
Content-Type: application/json
{
  "app": 232,
  "version": 3,
  "name": "0.3.3",
  "update": "2014-11-02T00:00:00.000Z",
  "changed": "1. New design application icon.\n 2. Fix some bugs. 3. Add GCM push supported",
  "url": "http://file.arkors.com/releases/demo-lastest.apk",
  "client": "Android",
  "compatible": "0.3.2 0.3.1 0.2.x 0.1.x"  
}
```

###### Status Codes:
* 200 – 更新 Update 记录成功
* 400 – Errors (invalid json, missing or invalid fields, etc)
* 404 - 没有找到版本的记录

##### DELETE /v1/updates/:app/:version

删除一个 Application 的版本信息。由 [Board](https://github.com/arkors/board) 模块通过 UI 界面进行调用。

###### Example Request:
```
DELETE /v1/updates/:id HTTP/1.1
Host: update.arkors.com
X-Arkors-Application-Log: 5024442115e7bd738354c1fac662aed5
X-Arkors-Application-Client: xxx.xx.xx.xxx,BOARD
Accept: application/json
```

###### Example Response
```
HTTP/1.1 200 OK
X-Arkors-Application-Log: cb21df532c6647383af7efa0fd8405f2
Content-Type: application/json
{
  "app": 232,
  "version": 3,
  "name": "0.3.3",
  "update": "2014-11-10T00:00:00.000Z",
  "changed": "1. New design application icon.\n 2. Fix some bugs.",
  "url": "http://file.arkors.com/releases/demo-lastest.apk",
  "client": "Android",
  "compatible": "0.3.2 0.3.1 0.2.x 0.1.x"    
}
```

###### Status Codes:
* 200 – 删除 Update 记录成功
* 400 – Errors (invalid json, missing or invalid fields, etc)
* 404 - 没有找到版本的记录
