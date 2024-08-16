<h1 align="center">说明</h1>

- PulsarMQ 架构与示例参考 [readme.pdf]

- 迁移代码参考该项目(本地测试通过)
   PulsarMQ SDK 接口简单易用，项目中 demo 加了一层封装暴露了基本的使用方式，牺牲了原始接口的灵活性。如果有较多定制化需求的请使用原始 SDK

## rabbitmq 测试说明

为了保持两种 MQ 在接口和概念上的一致，demo 中 rabbitmq 仅使用 `fanout` 一种场景。

| 名词         | pulsarMQ        | rabbitMQ           |
| ------------ | --------------- | ------------------ |
| subscription | 订阅            | 队列               |
| topic        | topic           | fanout 交换机      |
| 路由规则     | 通过 topic 路由 | 通过 exchange 路由 |

rabbitmq 使用 `fanout`类型交换机来模拟 pulsar 的 topic，所有发往此交换机的消息会被广播到与之绑定的 queue 上。同时，rabbitmq 用 `queue`模拟 pulsar 的订阅。
