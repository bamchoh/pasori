# pasori

This library allows you to get the IDm from pasori which is Felica card reader.

# install

`go get github.com/bamchoh/pasori`

# usage

To get the IDm, Vender ID and Product ID are needed. Before using, you need to get VID and PID from your Felica card reader.

For pasori, Vender ID is 0x054C, Product ID is 0x06C3.

```
package main

import (
	"fmt"
	"github.com/bamchoh/pasori"
)

var (
	VID uint16 = 0x054C // SONY
	PID uint16 = 0x06C3 // RC-S380
)

func main() {
	idm, err := pasori.GetID(VID, PID)
	if err != nil {
		panic(err)
	}
	fmt.Println(idm)
}
```

# license

MIT license

# references

To imprement this library, I referenced below article deeply. @ysomei, Thank you for your allowing me to refer your article.
* 今更ですが、SONY RC-S380 で Suica の IDm を読み込んでみた
https://qiita.com/ysomei/items/32f366b61a7b631c4750
