
1. Make a yaml config file myconfig.yml:

```yaml
# krb5: the local path of the krb5.conf
krb5: /etc/krb5.conf
client:
  # principal that used to login and request ticket
  # tss or tss@HADOOP.COM
  principal: tss
  # the local path of the keytab of the principal
  keytab: /home/tss/user.keytab
server:
  # optional: user defined server principal name, if not set, HTTP/hostname@realm will be used
  # principal:
  # backend proxy to
  upstream: http://some:port
  listen: :5000
```

2. build and run the program
```bash
go build

./spnego-proxy myconfig.yml
```

3. access the proxy (port 5000 by default)
```bash
curl http://localhost:5000
```
