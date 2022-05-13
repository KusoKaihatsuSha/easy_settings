# Source package for easy saving settings
Package for savings JSON settings to single-alone database (type Bolt) with use minimal counts of lines code.

### Add to use

```
$ go get github.com/KusoKaihatsuSha/easy_settings.git
```

### Usage example

In example will be created database `data.db` in same folder and create `data_group.json` as export settings from database

```go
package main

import (
	"fmt"

	es "github.com/KusoKaihatsuSha/easy_settings"
)

func main() {
	es.Pack("group").Item("test011").Add("id001", "001").Add("id002", "002")
	es.Pack("group").Item("test012").Add("id003", "003")
	es.Pack("group").Item("test021").Add("id004", "004")
	fmt.Println(es.Pack("group").Item("test011", true).Get("id001"))
	fmt.Println(es.Pack("group").Item("test01", true).Get("id", true))
	fmt.Println(es.Pack("group").Item("test", true).Get("id", true))
	es.Pack("group").Item("test", true).SaveJson()
}

```

### Output file if using `SaveJson()`

```json
{
  "ID": "4t81omdwfn8Am65c",
  "Name": "group",
  "Items": [
    {
      "Parent": null,
      "Name": "test011",
      "ID": "5pqsia7a5vC5lv67",
      "Values": [
        {
          "Key": "id001",
          "Value": "001"
        },
        {
          "Key": "id002",
          "Value": "002"
        }
      ]
    },
    {
      "Parent": null,
      "Name": "test012",
      "ID": "106UtcxTnn1w5Ux6",
      "Values": [
        {
          "Key": "id003",
          "Value": "003"
        }
      ]
    },
    {
      "Parent": null,
      "Name": "test021",
      "ID": "oeBj70fWn8gnjj65",
      "Values": [
        {
          "Key": "id004",
          "Value": "004"
        }
      ]
    }
  ],
  "Db": null,
  "Uniq": false
}
```

