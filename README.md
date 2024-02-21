Package httpx provides out-of-box HTTP client/server settings and implementations.

## Why we need a dedicated package for using HTTP?

Although Golang's builtin HTTP package (`net/http`) is battery included and highly performant,
we still need to tune a few knobs, for example:

- proper timeout settings in client & server to prevent exhausting server resource
- up-to-date SSL/TLS setup for enforcing secure transport

It's tedious to remember and set them right for these subtle settings.

## What's `httpx` approach?

To avoid repeating ourself, in `httpx`, we use **scenario based** approach to
provide different out-of-box setups for different kind of services, clients.

We expect the service builder to import this package and let it configure everything by default.

### But my application has configured many settings on the http server, will it be difficult to migrate?

`httpx` also provides adapter helpers for adapting existing implementations.

**server usage**

```diff
 package main

 import (
     "net/http"

     "hcp/toolkit/httpx/app"
 )

 func main() {
-    appServer := &http.Server{}
+    appServer, err := app.ApplyServer(&http.Server{})
+    if err != nil {
+        // handle err
+    }

     // ... continue to adapt other settings in the server instance
 }

```

## Wait, how can you assure your implementation is attack-proof?

For each attack surface, we provide (or plan to provide) a set of tests to ensure the implementation
can survive from the attack.

Below are the attack surfaces covered:

- [x] server: [slowloris][]
- [ ] server: [R.U.D.Y][]
- [x] server: slow body read

[slowloris]: https://www.cloudflare.com/learning/ddos/ddos-attack-tools/slowloris/
[R.U.D.Y]: https://www.cloudflare.com/learning/ddos/ddos-attack-tools/r-u-dead-yet-rudy/
