package main

import (
   "2a.pages.dev/googleplay"
   "flag"
   "fmt"
   "os"
)

type flags struct {
   device bool
   doc string
   email string
   googleplay.Client
   passwd string
   platform int64
   purchase bool
   single bool
   vc uint64
}

func main() {
   var f flags
   f.Client = googleplay.Default_Client
   flag.StringVar(&f.doc, "d", "", "doc")
   flag.BoolVar(&f.device, "device", false, "create device")
   flag.StringVar(&f.email, "email", "", "your email")
   flag.IntVar(&f.Log_Level, "log", f.Log_Level, "log level")
   flag.Int64Var(&f.platform, "p", 0, googleplay.Platforms.String())
   flag.StringVar(&f.passwd, "passwd", "", "password")
   flag.BoolVar(&f.purchase, "purchase", false, "purchase request")
   flag.BoolVar(&f.single, "s", false, "single APK")
   flag.Uint64Var(&f.vc, "v", 0, "version code")
   flag.Parse()
   dir, err := os.UserHomeDir()
   if err != nil {
      panic(err)
   }
   dir += "/googleplay"
   os.Mkdir(dir, os.ModePerm)
   if f.passwd != "" {
      err := f.do_auth(dir)
      if err != nil {
         panic(err)
      }
   } else {
      platform := googleplay.Platforms[f.platform]
      if f.device {
         err := f.do_device(dir, platform)
         if err != nil {
            panic(err)
         }
      } else if f.doc != "" {
         head, err := f.do_header(dir, platform)
         if err != nil {
            panic(err)
         }
         if f.purchase {
            err := f.Purchase(head, f.doc)
            if err != nil {
               panic(err)
            }
         } else if f.vc >= 1 {
            err := f.do_delivery(head)
            if err != nil {
               panic(err)
            }
         } else {
            detail, err := f.do_details(head)
            if err != nil {
               panic(err)
            }
            fmt.Println(string(detail))
         }
      } else {
         flag.Usage()
      }
   }
}
