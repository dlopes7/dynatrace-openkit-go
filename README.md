# Dynatrace Community OpenKit Go

This is a hobbyist implementation of the [Dynatrace Openkit](https://www.dynatrace.com/support/help/extend-dynatrace/dynatrace-sdks/openkit/) for Golang  

It is **not** officially supported by Dynatrace.

## Install

`go get github.com/dlopes7/dynatrace-openkit-go@v1.1.1`

## Sample

```go
package main

import (
	"fmt"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo"
	"math/rand"
)

func main() {
	openkit := openkitgo.NewOpenKitBuilder("https://tenant.app.url/mbeacon", "my-app-id", 19).
		WithApplicationName("Sample APp").
		WithApplicationVersion("1.000").
		WithManufacturer("Sample Inc").
		WithModelID("Model S").
		WithOperatingSystem("arch").
		Build()

	session := openkit.CreateSession("192.168.15.103")
	session.IdentifyUser(fmt.Sprintf("USER_%d", rand.Intn(10)))

	rootAction := session.EnterAction("My User Action")

	action := rootAction.EnterAction("My Child Action 1 ")
	action.LeaveAction()

	rootAction.EnterAction("My Child Action 2")

	rootAction.LeaveAction()
	session.End()

}

```

