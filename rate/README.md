### 令牌桶

token 生成速率 rate
桶容量 cap

计算需要补充的 token
supply_token = (now - last_time) * rate

目前 token 总数
tokens = min(tokens+supply_token, cap)

tokens -= reqTokens

更新 last_time 和 tokens