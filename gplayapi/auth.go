package gplayapi

import (
   "bytes"
   "errors"
   "fmt"
   "net/http"
   "net/url"
   "strconv"

   "github.com/Juby210/gplayapi-go/gpproto"
   "google.golang.org/protobuf/proto"
)

type _AuthData struct {
   _Email                         string
   _AASToken                      string
   _AuthToken                     string
   _GsfID                         string
   _DeviceCheckInConsistencyToken string
   _DeviceConfigToken             string
   _DFECookie                     string
   _Locale                        string
}

func (client *_GooglePlayClient) _GenerateGsfID() (gsfID string, err error) {
   req := client._DeviceInfo._GenerateAndroidCheckInRequest()
   checkInResp, err := client.checkIn(req)
   if err != nil {
      return
   }
   gsfID = fmt.Sprintf("%x", checkInResp.GetAndroidId())
   client._AuthData._GsfID = gsfID
   client._AuthData._DeviceCheckInConsistencyToken = checkInResp.GetDeviceCheckinConsistencyToken()
   return
}

func (client *_GooglePlayClient) checkIn(req *gpproto.AndroidCheckinRequest) (resp *gpproto.AndroidCheckinResponse, err error) {
   b, err := proto.Marshal(req)
   if err != nil {
      return
   }
   r, _ := http.NewRequest("POST", _UrlCheckIn, bytes.NewReader(b))
   client.setAuthHeaders(r)
   r.Header.Set("Content-Type", "application/x-protobuffer")
   r.Header.Set("Host", "android.clients.google.com")
   b, _, err = doReq(r)
   if err != nil {
      return
   }
   resp = &gpproto.AndroidCheckinResponse{}
   err = proto.Unmarshal(b, resp)
   return
}

func (client *_GooglePlayClient) uploadDeviceConfig() (*gpproto.UploadDeviceConfigResponse, error) {
   b, err := proto.Marshal(&gpproto.UploadDeviceConfigRequest{DeviceConfiguration: client._DeviceInfo._GetDeviceConfigProto()})
   if err != nil {
      return nil, err
   }
   r, _ := http.NewRequest("POST", _UrlUploadDeviceConfig, bytes.NewReader(b))
   payload, err := client.doAuthedReq(r)
   if err != nil {
      return nil, err
   }
   return payload.UploadDeviceConfigResponse, nil
}

func (client *_GooglePlayClient) _GenerateGPToken() (string, error) {
   params := &url.Values{}
   client.setDefaultAuthParams(params)
   client.setAuthParams(params)

   params.Set("app", "com.google.android.gms")
   params.Set("service", "oauth2:https://www.googleapis.com/auth/googleplay")

   r, _ := http.NewRequest("POST", _UrlAuth+"?"+params.Encode(), nil)
   client.setAuthHeaders(r)
   b, _, err := doReq(r)
   if err != nil {
      return "", nil
   }
   resp := parseResponse(string(b))
   token, ok := resp["Auth"]
   if !ok {
      return "", errors.New("authentication failed: could not generate oauth token")
   }
   return token, nil
}

func (client *_GooglePlayClient) setAuthHeaders(r *http.Request) {
   r.Header.Set("app", "com.google.android.gms")
   r.Header.Set("User-Agent", client._DeviceInfo._GetAuthUserAgent())
   if client._AuthData._GsfID != "" {
      r.Header.Set("device", client._AuthData._GsfID)
   }
}

func (client *_GooglePlayClient) setDefaultHeaders(r *http.Request) {
   data := client._AuthData
   r.Header.Set("Authorization", "Bearer "+data._AuthToken)
   r.Header.Set("User-Agent", client._DeviceInfo._GetUserAgent())
   r.Header.Set("X-DFE-Device-Id", data._GsfID)
   r.Header.Set("Accept-Language", "en-GB")
   r.Header.Set(
      "X-DFE-Encoded-Targets",
      "CAESN/qigQYC2AMBFfUbyA7SM5Ij/CvfBoIDgxHqGP8R3xzIBvoQtBKFDZ4HAY4FrwSVMasHBO0O2Q8akgYRAQECAQO7AQEpKZ0CnwECAwRrAQYBr9PPAoK7sQMBAQMCBAkIDAgBAwEDBAICBAUZEgMEBAMLAQEBBQEBAcYBARYED+cBfS8CHQEKkAEMMxcBIQoUDwYHIjd3DQ4MFk0JWGYZEREYAQOLAYEBFDMIEYMBAgICAgICOxkCD18LGQKEAcgDBIQBAgGLARkYCy8oBTJlBCUocxQn0QUBDkkGxgNZQq0BZSbeAmIDgAEBOgGtAaMCDAOQAZ4BBIEBKUtQUYYBQscDDxPSARA1oAEHAWmnAsMB2wFyywGLAxol+wImlwOOA80CtwN26A0WjwJVbQEJPAH+BRDeAfkHK/ABASEBCSAaHQemAzkaRiu2Ad8BdXeiAwEBGBUBBN4LEIABK4gB2AFLfwECAdoENq0CkQGMBsIBiQEtiwGgA1zyAUQ4uwS8AwhsvgPyAcEDF27vApsBHaICGhl3GSKxAR8MC6cBAgItmQYG9QIeywLvAeYBDArLAh8HASI4ELICDVmVBgsY/gHWARtcAsMBpALiAdsBA7QBpAJmIArpByn0AyAKBwHTARIHAX8D+AMBcRIBBbEDmwUBMacCHAciNp0BAQF0OgQLJDuSAh54kwFSP0eeAQQ4M5EBQgMEmwFXywFo0gFyWwMcapQBBugBPUW2AVgBKmy3AR6PAbMBGQxrUJECvQR+8gFoWDsYgQNwRSczBRXQAgtRswEW0ALMAREYAUEBIG6yATYCRE8OxgER8gMBvQEDRkwLc8MBTwHZAUOnAXiiBakDIbYBNNcCIUmuArIBSakBrgFHKs0EgwV/G3AD0wE6LgECtQJ4xQFwFbUCjQPkBS6vAQqEAUZF3QIM9wEhCoYCQhXsBCyZArQDugIziALWAdIBlQHwBdUErQE6qQaSA4EEIvYBHir9AQVLmgMCApsCKAwHuwgrENsBAjNYswEVmgIt7QJnN4wDEnta+wGfAcUBxgEtEFXQAQWdAUAeBcwBAQM7rAEJATJ0LENrdh73A6UBhAE+qwEeASxLZUMhDREuH0CGARbd7K0GlQo",
   )
   r.Header.Set(
      "X-DFE-Phenotype",
      "H4sIAAAAAAAAAB3OO3KjMAAA0KRNuWXukBkBQkAJ2MhgAZb5u2GCwQZbCH_EJ77QHmgvtDtbv-Z9_H63zXXU0NVPB1odlyGy7751Q3CitlPDvFd8lxhz3tpNmz7P92CFw73zdHU2Ie0Ad2kmR8lxhiErTFLt3RPGfJQHSDy7Clw10bg8kqf2owLokN4SecJTLoSwBnzQSd652_MOf2d1vKBNVedzg4ciPoLz2mQ8efGAgYeLou-l-PXn_7Sna1MfhHuySxt-4esulEDp8Sbq54CPPKjpANW-lkU2IZ0F92LBI-ukCKSptqeq1eXU96LD9nZfhKHdtjSWwJqUm_2r6pMHOxk01saVanmNopjX3YxQafC4iC6T55aRbC8nTI98AF_kItIQAJb5EQxnKTO7TZDWnr01HVPxelb9A2OWX6poidMWl16K54kcu_jhXw-JSBQkVcD_fPsLSZu6joIBAAA",
   )
   r.Header.Set("X-DFE-Client-Id", "am-android-google")
   r.Header.Set("X-DFE-Network-Type", "4")
   r.Header.Set("X-DFE-Content-Filters", "")
   r.Header.Set("X-Limit-Ad-Tracking-Enabled", "false")
   r.Header.Set("X-Ad-Id", "LawadaMera")
   r.Header.Set("X-DFE-UserLanguages", "en_GB")
   r.Header.Set("X-DFE-Request-Params", "timeoutMs=4000")

   if data._DeviceCheckInConsistencyToken != "" {
      r.Header.Set("X-DFE-Device-Checkin-Consistency-Token", data._DeviceCheckInConsistencyToken)
   }

   if data._DeviceConfigToken != "" {
      r.Header.Set("X-DFE-Device-Config-Token", data._DeviceConfigToken)
   }

   if data._DFECookie != "" {
      r.Header.Set("X-DFE-Cookie", data._DFECookie)
   }

   if client._DeviceInfo._SimOperator != "" {
      r.Header.Set("X-DFE-MCCMNC", client._DeviceInfo._SimOperator)
   }
}

func (client *_GooglePlayClient) setDefaultAuthParams(params *url.Values) {
   if client._AuthData._GsfID != "" {
      params.Set("androidId", client._AuthData._GsfID)
   }
   params.Set("sdk_version", strconv.Itoa(int(client._DeviceInfo._Build.GetSdkVersion())))
   params.Set("email", client._AuthData._Email)
   params.Set("google_play_services_version", strconv.Itoa(int(client._DeviceInfo._Build.GetGoogleServices())))
   params.Set("device_country", "us")
   params.Set("lang", "en-gb")
   params.Set("callerSig", "38918a453d07199354f8b19af05ec6562ced5788")
}

func (client *_GooglePlayClient) setAuthParams(params *url.Values) {
   params.Set("app", "com.android.vending")
   params.Set("client_sig", "38918a453d07199354f8b19af05ec6562ced5788")
   params.Set("callerPkg", "com.google.android.gms")
   params.Set("Token", client._AuthData._AASToken)
   params.Set("oauth2_foreground", "1")
   params.Set("token_request_options", "CAA4AVAB")
   params.Set("check_email", "1")
   params.Set("system_partition", "1")
}
