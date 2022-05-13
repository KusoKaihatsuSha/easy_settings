# Source package for easy saving settings
Package for savings JSON settings to single-alone database (type Bolt) as one line in code

### Usage example

```go
import es "github.com/KusoKaihatsuSha/easy_settings"

...

es.Pack("messages").Item("test", true)
es.Pack("messages").Item("test").Add("id", "value003")
es.Pack("messages").Item("test").Clear("value", true)

...

es.Pack("users").Item("testuser").Add("id", "value003")
es.Pack("users").Item("testuser", true).Get("id", true)
```

### Add to use

```
$ go get github.com/KusoKaihatsuSha/easy_settings.git
```

