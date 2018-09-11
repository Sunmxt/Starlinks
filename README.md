### 短链接

短链接服务

***Current state: Prototype. Maintained but Unstable.*** *Be cautious to use in production.*

```bash
make
```

---

#### API 参考

API 返回的 error_code ，意义参考下表：

| 值   | 错误描述 |
| ---- | -------- |
| 0    | 成功     |
| 0101 | 参数错误 |
| 0500 | 未知错误 |





-  增加短连接

  - **POST  http://.../v1/links**

    | 参数名 | 参数类型 | 参数描述 | 示例值                                        |
    | ------ | -------- | -------- | --------------------------------------------- |
    | URL    | string   | 链接     | http://www.starstudio.org/opensource/starlink |

  - **返回**：

    ```JSON
    {
        "error_code": <错误码>,
        "error_desc": <简要的错误描述>,
        "short_route": <短链接>, // string, 发生错误时，为空
        "id": <短链接ID>, // string, 十进制数字，发生错误时，为空
    }
    ```

