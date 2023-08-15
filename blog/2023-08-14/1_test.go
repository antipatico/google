package play

import (
   "154.pages.dev/http/option"
   "fmt"
   "testing"
)

func Test_Embedded_Setup(t *testing.T) {
   option.No_Location()
   option.Verbose()
   setup, err := new_embedded_setup()
   if err != nil {
      t.Fatal(err)
   }
   fmt.Println(host_gaps(setup.cookies))
   fmt.Println()
   fmt.Println(setup.identifier())
}
