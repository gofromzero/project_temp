# TODO
1. [项目规范](https://github.com/elsewhencode/project-guidelines)
2. 如果可以的话最好使用 Docker 镜像。
3. [api 设计指南](https://cloud.google.com/apis/design)
4. 定义公共Code
```
200 OK GET, PUT 或 POST 请求响应成功.

201 Created 标识一个新实例创建成功。当创建一个新的实例，请使用POST方法并返回201状态码。

304 Not Modified 发现资源已经缓存在本地，浏览器会自动减少请求次数。

400 Bad Request 请求未被处理，因为服务器不能理解客户端是要什么。

401 Unauthorized 因为请求缺少有效的凭据，应该使用所需的凭据重新发起请求。

403 Forbidden 意味着服务器理解本次请求，但拒绝授权。

404 Not Found 表示未找到请求的资源。

500 Internal Server Error 表示请求本身是有效，但由于某些意外情况，服务器无法实现，服务器发生了故障。
```
5. mysql 用户分离