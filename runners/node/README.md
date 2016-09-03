```
$ base64 testExample.js 
dmFyIGIgPSAzCgpmb3IoaT0wOyBpPGI7IGkrKykgewoJY29uc29sZS5sb2coImp1c3QgYSBwcmludGxpbmUiKQp9Cgpjb25zb2xlLmxvZygndGhhbmtzIGZvciB3YXRjaGluZycpCg==
```

```
$ curl -XPOST -d 'dmFyIGIgPSAzCgpmb3IoaT0wOyBpPGI7IGkrKykgewoJY29uc29sZS5sb2coImp1c3QgYSBwcmludGxpbmUiKQp9Cgpjb25zb2xlLmxvZygndGhhbmtzIGZvciB3YXRjaGluZycpCg==' localhost:8080/node
{"output":"anVzdCBhIHByaW50bGluZQpqdXN0IGEgcHJpbnRsaW5lCmp1c3QgYSBwcmludGxpbmUKdGhhbmtzIGZvciB3YXRjaGluZwo="}%  
```

```
$ echo "anVzdCBhIHByaW50bGluZQpqdXN0IGEgcHJpbnRsaW5lCmp1c3QgYSBwcmludGxpbmUKdGhhbmtzIGZvciB3YXRjaGluZwo=" | base64 --decode
just a printline
just a printline
just a printline
thanks for watching
```
