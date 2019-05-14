package main

import (
    "fmt"
    "os"
    "time"
    "log"
    "bufio"
    "net/http"
    "io/ioutil"
	"github.com/mkideal/cli"
	"github.com/xrash/smetrics"
	"github.com/badoux/goscraper"
	"github.com/fatih/color"
	"sync"
	"strings"
	"strconv"
	"sort"
	"crypto/tls"
	"io"
	"sync/atomic"
	"net/url"
	"github.com/hashicorp/go-version"
	"os/signal"
	"syscall"
	"github.com/ti/nasync"
    "database/sql"
    "path/filepath"
    _ "github.com/mattn/go-sqlite3"
)


var mu sync.Mutex
var wg2 sync.WaitGroup

var workersCount = 2
var timeoutSec = 5
var verboseMode = false
var intelligentMode = false
var CMSmode = false
var SpreadMode = false
var Statuscode = 0
var Excludecode = 0
var currentCount int = 0 
var currentCount1 int = 0
var ContinueNum int = 0 
var proxyMode = false
var enableDebug = false
var lookupMode = false

var totalListCount int = 0
var currentListCount int = 1
var currentListCount1 int = 0

var currentFakeCount int32 = 0 
var currentProgressCount int32 = 0 

var Pathsource = ""
var tmpTitleList [][]string	
var tmpResultList [][]string	
var tmpResultList1 []string	
var tmpResultList4 []string

var completedPathList []string
var completedCount = 0
var tmpFoundList [] string

var wpFileList []string
var joomlaFileList []string	
var drupalFileList []string
var proxy_addr=""
var reachedTheEnd=false
var reachedTheEnd1=false

var whitelistList []string
var blacklistList []string

var identUser = ""
var identPass = ""

var userAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.27 Safari/537.36"
        
func f(from string) {
    for i := 0; i < 3; i++ {
        fmt.Println(from, ":", i)
    }
}


func lookupURI(searchTerm string) ([]string) {
	var results []string	
	ex, err := os.Executable()
	if err == nil {
		exPath := filepath.Dir(ex)
		var pFilename=exPath+"/pathbrute.sqlite"
		_, err1 := os.Stat(pFilename)
		if os.IsNotExist(err1) {
			fmt.Printf("[*] Database file: %s not exists\n", pFilename)
			os.Exit(3)
		} else {
			database, _ := sql.Open("sqlite3", pFilename)
			rows, _ := database.Query("SELECT field1,field2,field3,field4 FROM db WHERE field3=='"+searchTerm+"'")
			var dataSource string
			var filename string
			var uriPath string
			var category string
			for rows.Next() {
				rows.Scan(&dataSource,&filename,&uriPath,&category)
				fmt.Println(dataSource+"\t"+filename+"\t"+uriPath+"\t"+category)
			}
		}
	}
	return results
}

func getRemoteSize(url string) (int64) {
	 resp, err := http.Head(url)
	 if err != nil {
			 fmt.Println("[-] Please check your Internet connection. Unable to connect to raw.githubusercontent.com.")
			 os.Exit(1)
	 }
	 if resp.StatusCode != http.StatusOK {
			 fmt.Println(resp.Status)
			 os.Exit(1)
	 }
	 size, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
	 downloadSize := int64(size)
	 return downloadSize
}
 func getFileSize(filename string) (int64){
	file, err := os.Open(filename)
	  if err != nil {
		  return 0
	  }
	  defer file.Close()
	  stat, err := file.Stat()
	  if err != nil {
		 return 0
	  }
	  var bytes int64
  	  bytes = stat.Size()
	  return bytes
 }
 func checkAndUpdate(downloadUrl string) (bool) {
	 tokens := strings.Split(downloadUrl,"/")
	 fileName := tokens[len(tokens)-1]
	 _, err1 := os.Stat(fileName)
	 if os.IsNotExist(err1) {
		fmt.Println("[+] Downloading: "+downloadUrl)
		err := DownloadFile(fileName, downloadUrl)
		if err!=nil {
			fmt.Println("[*] Error: ",err)
		} else {
			return true
		}
	 } else {
		 var localFileSize=getFileSize(fileName)		
		 downloadSize:=getRemoteSize(downloadUrl)
		 if (localFileSize!=downloadSize) {
			fmt.Println("[+] Downloading: "+downloadUrl)
			err := DownloadFile(fileName, downloadUrl)
			if err!=nil {
				fmt.Println("[*] Error: ",err)
			} else {
				return true
			}
		 } else {
		 	return false
		 }	
	} 
	return false
}
	
 
func checkWebsite(urlChan chan string) (string,bool) {
	var websiteUp=false
	var currentURL=""
	for newURL1 := range urlChan {
		currentURL=newURL1
		var userAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.27 Safari/537.36"

		timeout := time.Duration(time.Duration(timeoutSec) * time.Second)
		client := http.Client{
			Timeout: timeout,
			CheckRedirect: redirectPolicyFunc,
		}
		if proxyMode==true {
			url_i := url.URL{}
			url_proxy, _ := url_i.Parse(proxy_addr)
			http.DefaultTransport.(*http.Transport).Proxy = http.ProxyURL(url_proxy)
		}
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		req, err := http.NewRequest("GET", newURL1, nil)
		req.SetBasicAuth(identUser, identPass)

		if err==nil {
			if (len(identUser) + len(identPass) > 0) {
				req.SetBasicAuth(identUser, identPass)
				req.Header.Add("User-Agent", userAgent)

			}
			req.Header.Add("User-Agent", userAgent)
			resp, err := client.Do(req)			
			if resp!=nil{					
				defer resp.Body.Close()
			}
			if err==nil{	
				fmt.Println(newURL1+" "+strconv.Itoa(resp.StatusCode))		
				websiteUp=true		
			} 
		}
		return newURL1,websiteUp
	}
	return currentURL,websiteUp
}
func checkStatusCode(statusCode int) (bool) {
	if (statusCode!=403 && statusCode!=301 && statusCode!=301 && statusCode!=503 && statusCode!=404 && statusCode!=406 && statusCode!=400 && statusCode!=500 && statusCode!=204 && statusCode!=302) {
		return true
	} else {
		return false
	}
}
func getPageBody(tmpUrl string) (string) {
	var pageBody=""
	timeout := time.Duration(time.Duration(timeoutSec) * time.Second)
	client := http.Client{
		Timeout: timeout,
		CheckRedirect: redirectPolicyFunc,
	}
	if proxyMode==true {
		url_i := url.URL{}
		url_proxy, _ := url_i.Parse(proxy_addr)
		http.DefaultTransport.(*http.Transport).Proxy = http.ProxyURL(url_proxy)
	}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	req, err := http.NewRequest("GET", tmpUrl, nil)
	if err==nil {
		req.SetBasicAuth(identUser, identPass)
		req.Header.Add("User-Agent", userAgent)
		resp, err := client.Do(req)		
		if resp!=nil{					
			defer resp.Body.Close()
		}
		if err == nil {
			body, err := ioutil.ReadAll(resp.Body)
			if err==nil {
				pageBody = string(body)
			}
		}
	}
	return pageBody
}

func cleanup() {
    fmt.Println("\nCTRL-C (interrupt signal)")
    /*
	for _, v := range tmpResultList {
		if !stringInSlice(v[0],tmpResultList1) {
			tmpResultList1 = append(tmpResultList1, v[0])
		}
	}
	
	var tmpResultList2 []string
	sort.Strings(tmpResultList1)
	for _, v := range tmpResultList1 {
		u, err := url.Parse(v)
		if err==nil {
			if len(u.Path)>0 {
				tmpResultList2 = append(tmpResultList2,v)
			}
		}
	}    
	var tmpResultList3 []string
	if len(tmpResultList2)>0 {
		for _, v := range tmpResultList2 {
			timeout := time.Duration(time.Duration(timeoutSec) * time.Second)
			client := http.Client{
				Timeout: timeout,
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},					
			}
			if proxyMode==true {
				url_i := url.URL{}
				url_proxy, _ := url_i.Parse(proxy_addr)
				http.DefaultTransport.(*http.Transport).Proxy = http.ProxyURL(url_proxy)
			}
			http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

			req, err := http.NewRequest("GET", v, nil)
			if err==nil {
				req.Header.Add("User-Agent", userAgent)
				resp, err := client.Do(req)		
				if resp!=nil{					
					defer resp.Body.Close()
				}
				if err == nil {
					s, err := goscraper.Scrape(v, 5)
					if err == nil {
						var tmpTitle=strings.TrimSpace(s.Preview.Title)
						var lenBody = 0
						body, err := ioutil.ReadAll(resp.Body)
						if err==nil {
							lenBody = len(body)
						}
						if checkStatusCode(resp.StatusCode)==true {
							var a = v+" ["+(strconv.Itoa(resp.StatusCode))+"] ["+strconv.Itoa(lenBody)+"] ["+tmpTitle+"]"
							tmpResultList3 = append(tmpResultList3,a)
						}
					}
				}
			}
		}
	}
	if len(tmpResultList3)>0 {
		fmt.Printf("\n")
		log.Printf("\n")

		var wg sync.WaitGroup
		urlChan := make(chan string)
		wg.Add(workersCount)
		for i := 0; i < workersCount; i++ {
			go func() {
				checkURL(urlChan)
				wg.Done()
			}()
		}
		for _, each := range tmpResultList3 {
			fmt.Println("b "+each)
			urlChan <- each
		}
		close(urlChan)  
		wg.Wait()		
	
	} else {
		fmt.Println("\n[*] No results found")
	}
	for {	
		if reachedTheEnd1==true {
			break
		}
		if int(currentListCount1)>=len(tmpResultList2) {
			reachedTheEnd1=true
		} 
	}
	fmt.Println("cleanup end")
	os.Exit(3)
	*/
	for _, v := range tmpResultList {
		if !stringInSlice(v[0],tmpResultList1) {
			tmpResultList1 = append(tmpResultList1, v[0])
		}
	}

	var tmpResultList2 []string	

	sort.Strings(tmpResultList1)
	for _, v := range tmpResultList1 {
		u, err := url.Parse(v)
		if err==nil {
			if len(u.Path)>0 {
				tmpResultList2 = append(tmpResultList2,v)
			}
		}
	}					
	
	if len(tmpResultList2)<1 {
		fmt.Println("\n[*] No results found")
		log.Printf("\n[*] No results found")
	} else {
		//time.Sleep(5 * time.Second)
		fmt.Printf("\n")
		log.Printf("\n")

		var wg sync.WaitGroup
		urlChan := make(chan string)
		wg.Add(workersCount)
		for i := 0; i < workersCount; i++ {
			go func() {	
				checkURL(urlChan)
				wg.Done()
			}()
		}		
		for _, each := range tmpResultList2 {
			//async.Do(checkURL1,each)
			urlChan <- each
		}
		close(urlChan)  
		wg.Wait()				
	}
	for {	
		if reachedTheEnd1==true {
			break
		}
		if int(currentListCount1)>=len(tmpResultList2) {
			reachedTheEnd1=true
		} 
	}

}

func removeCharacters(input string, characters string) string {
	 filter := func(r rune) rune {
		 if strings.IndexRune(characters, r) < 0 {
				 return r
		 }
		 return -1
	 }
	 return strings.Map(filter, input)
}

func redirectPolicyFunc(req *http.Request, via []*http.Request) error{
	if (len(identUser) + len(identPass)) > 0 {
		req.SetBasicAuth(identUser, identPass)
	}
	return http.ErrUseLastResponse
}

func getPage(newURL1 string) (string, string, int, int) {
	var tmpStatusCode = 0
	var tmpTitle = ""
	var lenBody=0
	var tmpFinalURL =""
	timeout := time.Duration(time.Duration(timeoutSec) * time.Second)
	client := http.Client{
		Timeout: timeout,
		CheckRedirect: redirectPolicyFunc,		
	}
	if proxyMode==true {
		url_i := url.URL{}
		url_proxy, _ := url_i.Parse(proxy_addr)
		http.DefaultTransport.(*http.Transport).Proxy = http.ProxyURL(url_proxy)
	}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	req, err := http.NewRequest("GET", newURL1, nil)
	if err==nil {
		req.Header.Add("User-Agent", userAgent)
		req.SetBasicAuth(identUser, identPass)
		resp, err := client.Do(req)		
		if resp!=nil{					
			defer resp.Body.Close()
		}
		if err==nil{					
			tmpStatusCode = resp.StatusCode
			s, err := goscraper.Scrape(newURL1, 5)			
			if err==nil {
				tmpTitle = s.Preview.Title
				tmpTitle=strings.TrimSpace(tmpTitle)
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err==nil {
				lenBody = len(body)
			}
			tmpFinalURL = resp.Request.URL.String()			
			return tmpFinalURL,tmpTitle,tmpStatusCode, lenBody
		} else {
			if strings.Contains(err.Error(),"server gave HTTP response to HTTPS client") {
				tmpStatusCode = 302
			} else if strings.Contains(err.Error(),"TLS handshake timeout") {
				tmpStatusCode = 408
			}
		}
	}
	return tmpFinalURL,tmpTitle,tmpStatusCode, lenBody
}

func testFakePath(urlChan chan string) {
    for newUrl := range urlChan {
		var getTmpTitle=""
		var getTmpStatusCode=0
		var getLenBody=0
		var getTmpFinalURL=""

		u, err := url.Parse(newUrl)
		if err != nil {
			panic(err)
		}
		var newURL1=u.Scheme+"://"+u.Host+"/NonExistence"

		getTmpFinalURL,getTmpTitle,getTmpStatusCode,getLenBody=getPage(newURL1)	

		if stringInSlice(getTmpTitle,whitelistList) {
			blacklistList = append(blacklistList, u.Scheme+"://"+u.Host)
		}

		newUrl = strings.Replace(newUrl, "/NonExistence", "", -1)
		if strings.HasSuffix(getTmpFinalURL,"/") {
			getTmpFinalURL=getTmpFinalURL[0:len(getTmpFinalURL)-1]
		}			
		if verboseMode==true {
			if getTmpStatusCode==200{
				if verboseMode==true {
					fmt.Printf("%s [%s] [%s] [%s]\n",newUrl+"/NonExistence", color.BlueString(strconv.Itoa(getTmpStatusCode)),strconv.Itoa(getLenBody), getTmpTitle)
					log.Printf("%s [%s] [%s] [%s]\n",newUrl+"/NonExistence", color.BlueString(strconv.Itoa(getTmpStatusCode)),strconv.Itoa(getLenBody), getTmpTitle)
				}
				var a = [][]string{{newUrl, strconv.Itoa(getTmpStatusCode), "",""}}
				tmpResultList = append(tmpResultList,a...)
			} else if getTmpStatusCode==401{
				if verboseMode==true {
					fmt.Printf("%s [%s] [%s] [%s]\n",newUrl+"/NonExistence", color.GreenString(strconv.Itoa(getTmpStatusCode)),strconv.Itoa(getLenBody), getTmpTitle)
					log.Printf("%s [%s] [%s] [%s]\n",newUrl+"/NonExistence", color.GreenString(strconv.Itoa(getTmpStatusCode)),strconv.Itoa(getLenBody), getTmpTitle)
				}
				var a = [][]string{{newUrl, strconv.Itoa(getTmpStatusCode), "",""}}
				tmpResultList = append(tmpResultList,a...)
			} else {
				if verboseMode==true {
					if getTmpStatusCode==0 {
						fmt.Printf("%s [%s] [%s] [%s]\n",newUrl+"/NonExistence", color.RedString(""),strconv.Itoa(getLenBody), getTmpTitle)
						log.Printf("%s [%s] [%s] [%s]\n",newUrl+"/NonExistence", color.RedString(""),strconv.Itoa(getLenBody), getTmpTitle)
					} else {
						fmt.Printf("%s [%s] [%s] [%s]\n",newUrl+"/NonExistence", color.RedString(strconv.Itoa(getTmpStatusCode)),strconv.Itoa(getLenBody), getTmpTitle)
						log.Printf("%s [%s] [%s] [%s]\n",newUrl+"/NonExistence", color.RedString(strconv.Itoa(getTmpStatusCode)),strconv.Itoa(getLenBody), getTmpTitle)
					}
				}
			}
			var a = [][]string{{newUrl, getTmpTitle, strconv.Itoa(getLenBody), strconv.Itoa(getTmpStatusCode)}}
			tmpTitleList = append(tmpTitleList,a...)
			//_ = a
		}

		atomic.AddInt32(&currentFakeCount, 1)
    }
}
func pathPrediction(newUrl string, statusCode int) (string) {
	var newPath=""
	var tmpFinalPath=""
	var finalURL=""
	var voidPath=false
	u, err := url.Parse(newUrl)
	if err!=nil {
		fmt.Println(err)
	}
	if err==nil {
		result := strings.Split(u.Path, "/")		
		if len(u.Path)>0 {
			for tmpCount := 1; tmpCount < len(result); tmpCount++ {
			// for i := range result {
				newPath=strings.Replace(u.Path,result[tmpCount],result[tmpCount]+"xx",1)
				_,_,tmpStatusCode, _:=getPage(u.Scheme+"://"+u.Host+newPath)
				//if enableDebug==true {
				//	fmt.Println("[debug1] "+newUrl+" -> "+newPath+" ["+strconv.Itoa(tmpStatusCode)+"]")					
				//}
				if statusCode!=tmpStatusCode {
					tmpFinalPath=tmpFinalPath+"/"+result[tmpCount]
					var newUrl1=u.Scheme+"://"+u.Host+tmpFinalPath
					if newUrl1!=newUrl {
						_,_,getTmpStatusCode1,_:=getPage(newUrl1)
						if enableDebug==true {
							fmt.Println("[debug2] "+newUrl+" -> "+tmpFinalPath+" ["+strconv.Itoa(getTmpStatusCode1)+"]")					
						}
					}
					if !strings.HasSuffix(newUrl,"/") {
						var newUrl4=u.Scheme+"://"+u.Host+tmpFinalPath+"xxx"
						_,_,getTmpStatusCode2,_:=getPage(newUrl4)
						if enableDebug==true {
							fmt.Println("[debug3] "+newUrl+" -> "+tmpFinalPath+"xxx ["+strconv.Itoa(getTmpStatusCode2)+"]")
						}
						if statusCode!=getTmpStatusCode2 {
							if checkStatusCode(getTmpStatusCode2)==true {
								//if (getTmpStatusCode2!=403 && getTmpStatusCode2!=503 && getTmpStatusCode2!=404 && getTmpStatusCode2!=406 && getTmpStatusCode2!=400 && getTmpStatusCode2!=500 && getTmpStatusCode2!=204 && getTmpStatusCode2!=302) {
								if enableDebug==true {
									fmt.Println("[predict1]")
								}
								finalURL=u.Scheme+"://"+u.Host+tmpFinalPath
							}
							//here
						}
						if !strings.Contains( u.Path,".") {					
							newUrl4=u.Scheme+"://"+u.Host+tmpFinalPath+"/xxx"
							_,_,getTmpStatusCode2,_=getPage(newUrl4)
							if enableDebug==true {
								fmt.Println("[debug4] "+newUrl+" -> "+tmpFinalPath+"/xxx ["+strconv.Itoa(getTmpStatusCode2)+"]")
							}
							if statusCode==getTmpStatusCode2 {
								if checkStatusCode(getTmpStatusCode2)==true {
									if enableDebug==true {
										fmt.Println("[predict2]")
									}
									finalURL=u.Scheme+"://"+u.Host
								}
							}
						}						
					} 
				} else {
					tmpFinalPath=tmpFinalPath+"/"+result[tmpCount]
					var newUrl1=u.Scheme+"://"+u.Host+tmpFinalPath
					_,_,getTmpStatusCode1,getLenBody1:=getPage(newUrl1)
					if enableDebug==true {
						if newUrl!=newUrl1 {
							fmt.Println("[debug5] "+newUrl+" -> "+tmpFinalPath+" ["+strconv.Itoa(getTmpStatusCode1)+"]")
							if statusCode==getTmpStatusCode1 && statusCode==200 {
								_,_,_,getLenBody2:=getPage(newUrl)
								if getLenBody1!=getLenBody2 {
										if enableDebug==true {									
											fmt.Println("[predict3]")
										}
										finalURL=newUrl
								}
							}
						}
					}
					if strings.Contains( u.Path,".") {					
						//////
						var newUrl4=u.Scheme+"://"+u.Host+tmpFinalPath+"xxx"
						_,_,getTmpStatusCode2,_:=getPage(newUrl4)
						if enableDebug==true {
							fmt.Println("[debug6] "+newUrl+" -> "+tmpFinalPath+"xxx ["+strconv.Itoa(getTmpStatusCode2)+"]")
						}
						if statusCode==getTmpStatusCode2 {
							if checkStatusCode(getTmpStatusCode2)==true {
								var s1=getPageBody(newUrl)
								var s2=getPageBody(newUrl4)
								if (smetrics.JaroWinkler(s1, s2, 0.7, 4))>0.7 {
									if enableDebug==true {
										fmt.Println("[predict4]")
									}
									finalURL=""
									voidPath=true
								} else {
									finalURL=u.Scheme+"://"+u.Host
								} 
							}					
						}
						newUrl4=u.Scheme+"://"+u.Host+tmpFinalPath+"/xxx"
						_,_,getTmpStatusCode2,_=getPage(newUrl4)
						if enableDebug==true {
							fmt.Println("[debug8] "+newUrl+" -> "+tmpFinalPath+"/xxx ["+strconv.Itoa(getTmpStatusCode2)+"]")
						}
						if statusCode==getTmpStatusCode2 {
							if checkStatusCode(getTmpStatusCode2)==true {
								var s1=getPageBody(newUrl)
								var s2=getPageBody(newUrl4)
								if (smetrics.JaroWinkler(s1, s2, 0.7, 4))>0.7 {
									if enableDebug==true {
										fmt.Println("[predict5]")
									}
									finalURL=""
									voidPath=true								
								}else {
									finalURL=u.Scheme+"://"+u.Host
								} 
							}
						}
					} else {
						var newUrl4=u.Scheme+"://"+u.Host+tmpFinalPath+"/xxx/"
						_,_,getTmpStatusCode2,_:=getPage(newUrl4)
						if enableDebug==true {
							fmt.Println("[debug10] "+newUrl+" -> "+tmpFinalPath+"/xxx/ ["+strconv.Itoa(getTmpStatusCode2)+"]")
						}
						if statusCode==getTmpStatusCode2 {
							if checkStatusCode(getTmpStatusCode2)==true {
								var s1=getPageBody(newUrl)
								var s2=getPageBody(newUrl4)
								if (smetrics.JaroWinkler(s1, s2, 0.7, 4))>0.7 {								
									if enableDebug==true {
										fmt.Println("[predict6]")
									}
									finalURL=""
									voidPath=true
								}else {
									finalURL=u.Scheme+"://"+u.Host
								} 
							}
						}
						if enableDebug==true {
							fmt.Println("[debug11] "+newUrl+" -> "+tmpFinalPath+" ["+strconv.Itoa(getTmpStatusCode1)+"]")	
						}
						newUrl4=u.Scheme+"://"+u.Host+tmpFinalPath+"/xxx/"
						_,_,getTmpStatusCode2,_=getPage(newUrl4)
						if enableDebug==true {
							fmt.Println("[debug12] "+newUrl+" -> "+tmpFinalPath+"/xxx/ ["+strconv.Itoa(getTmpStatusCode2)+"]")
						}
						if statusCode==getTmpStatusCode2 {
							if checkStatusCode(getTmpStatusCode2)==true {
								var s1=getPageBody(newUrl)
								var s2=getPageBody(newUrl4)
								if (smetrics.JaroWinkler(s1, s2, 0.7, 4))>0.7 {								
									if enableDebug==true {
										fmt.Println("[predict7]")									
									}
									finalURL=""
									voidPath=true
								} else {
									finalURL=u.Scheme+"://"+u.Host
								} 
							}
						}
					}	
					if getTmpStatusCode1!=statusCode{
						lenCount1 := strings.Count(tmpFinalPath, "/")
						if !strings.HasSuffix(u.Path,"/") && lenCount1<2{
							lenCount2 := strings.Count(u.Path, ".")
							if lenCount2>0 {							
								var newUrl2=u.Scheme+"://"+u.Host+"/xxx.jsp"
								_,_,getTmpStatusCode2,_:=getPage(newUrl2)
								if enableDebug==true {
									fmt.Println("[debug13] "+newUrl+" -> /xxx.jsp ["+strconv.Itoa(getTmpStatusCode2)+"]")
								}
								if getTmpStatusCode2==statusCode {
									if enableDebug==true {
										fmt.Println("[predict8]")
									}
									finalURL=u.Scheme+"://"+u.Host
								}
								var newUrl3=u.Scheme+"://"+u.Host+"/NonExistence/xxx.jsp"
								_,_,getTmpStatusCode3,_:=getPage(newUrl3)
								if enableDebug==true {
									fmt.Println("[debug14] "+newUrl+" -> /NonExistence/xxx.jsp ["+strconv.Itoa(getTmpStatusCode3)+"]")
								}
						
								if getTmpStatusCode1==getTmpStatusCode2 && statusCode==getTmpStatusCode3{
									if checkStatusCode(getTmpStatusCode2)==true {
										if enableDebug==true {
											fmt.Println("[predict9]")	
										}
										finalURL=u.Scheme+"://"+u.Host
									}
								}
							} 
						} else {				
							lenCount1 := strings.Count(tmpFinalPath, "/")	
							if lenCount1<1 {
								lenCount2 := strings.Count(tmpFinalPath, ".")	
								var newUrl2=""
								if lenCount2>0 {
									newUrl2=u.Scheme+"://"+u.Host+"/xxx.jsp"
								} else {
									newUrl2=u.Scheme+"://"+u.Host+"/xxx"
								}
								_,_,getTmpStatusCode2,_:=getPage(newUrl2)
								if getTmpStatusCode1==getTmpStatusCode2 {
									if checkStatusCode(getTmpStatusCode2)==true {						
										fmt.Println("[predict10]")
										finalURL=u.Scheme+"://"+u.Host
									}
								}
								if getTmpStatusCode2==statusCode {
									if enableDebug==true {
										fmt.Println("[predict11]")
									}
									finalURL=u.Scheme+"://"+u.Host
								}
							} 
						}
					}
				}
				if tmpCount==len(result)-1 {
					break
				}
			}
		}
	}
	var result=""
	if voidPath==false {
		if len(finalURL)>0 {
			if finalURL==u.Scheme+"://"+u.Host {
				if enableDebug==true {
					fmt.Println("[predict12]")
				}
				result=newUrl
			} else {
				if enableDebug==true {
					fmt.Println("[predict13]")
				}
				result=finalURL
			}
		} else {
			if enableDebug==true {
				fmt.Println("[predict14]")
			}
			result=newUrl
		}
	}
	return result
}
func addToCompleteList(newUrl string) {
	//completedPathList=append(completedPathList,newUrl)
	completedCount+=1
}

func checkURL1(v string) {
	var tmpResultList3 []string
	timeout := time.Duration(time.Duration(timeoutSec) * time.Second)
	client := http.Client{
		Timeout: timeout,
		CheckRedirect: redirectPolicyFunc,				
	}
	if proxyMode==true {
		url_i := url.URL{}
		url_proxy, _ := url_i.Parse(proxy_addr)
		http.DefaultTransport.(*http.Transport).Proxy = http.ProxyURL(url_proxy)
	}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	req, err := http.NewRequest("GET", v, nil)
	req.Header.Add("User-Agent", userAgent)
	req.SetBasicAuth(identUser, identPass)
	resp, err := client.Do(req)		
	if resp!=nil{					
		defer resp.Body.Close()
	}

	s, err := goscraper.Scrape(v, 5)
	var lenBody = 0
	var tmpTitle = ""
	if err==nil {
		tmpTitle=strings.TrimSpace(s.Preview.Title)						
		body, err3 := ioutil.ReadAll(resp.Body)
		if err3==nil {
			lenBody = len(body)
		}
		if checkStatusCode(resp.StatusCode)==true {
			if intelligentMode==false {
				if resp.StatusCode==401 {
					tmpTitle=strings.Replace(tmpTitle,"\n"," ",1)
					fmt.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",v, color.GreenString(strconv.Itoa(resp.StatusCode)),  lenBody, tmpTitle)								
					log.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",v, color.GreenString(strconv.Itoa(resp.StatusCode)),  lenBody, tmpTitle)
				} else if (resp.StatusCode==200) {
					tmpTitle=strings.Replace(tmpTitle,"\n"," ",1)
					fmt.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",v, color.BlueString(strconv.Itoa(resp.StatusCode)),  lenBody, tmpTitle)								
					log.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",v, color.BlueString(strconv.Itoa(resp.StatusCode)),  lenBody, tmpTitle)
				} else {
					tmpTitle=strings.Replace(tmpTitle,"\n"," ",1)
					fmt.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",v, color.RedString(strconv.Itoa(resp.StatusCode)),  lenBody, tmpTitle)								
					log.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",v, color.RedString(strconv.Itoa(resp.StatusCode)),  lenBody, tmpTitle)
				}				
			} else {
				var initialStatusCode=resp.StatusCode
				var initialPageSize=lenBody

				u, err := url.Parse(v)
				if err != nil {
					panic(err)
				}
				numberOfa := strings.Count(u.Path, "/")				
				tmpSplit2 :=strings.Split(u.Path,"/")
				var counter1=numberOfa
				if numberOfa<3 {	
					// var newURL3=u.Scheme+"://"+u.Host
					// getTmpFinalURL3,getTmpTitle3,getTmpStatusCode3,getLenBody3:=getPage(newURL3)
					var getTmpStatusCode4=0
					var getLenBody4=0
					var newURL4=""
				
					if !strings.HasSuffix(v,"/") {
						if strings.Contains( u.Path,".") {	
							tmpSplit1 :=strings.Split(u.Path,".")
							var fileExt = ("."+tmpSplit1[len(tmpSplit1)-1])

							tmpSplit3 :=strings.Split(u.Path,fileExt)
							newURL4=u.Scheme+"://"+u.Host+tmpSplit3[0]+"xxx"+fileExt
							_,_,getTmpStatusCode4,getLenBody4=getPage(newURL4)
							if enableDebug==true {
								fmt.Println("[debug15] "+newURL4+" ["+strconv.Itoa(getTmpStatusCode4)+"]")
							}
							if (resp.StatusCode==getTmpStatusCode4 && (getTmpStatusCode4!=200 && getTmpStatusCode4!=401 && getTmpStatusCode4!=405)) {
								var returnURL=u.Scheme+"://"+u.Host
								tmpResultList3 = append(tmpResultList3, returnURL)
							} else {
								if lenBody!=getLenBody4 {
									var returnURL=v
									tmpResultList3 = append(tmpResultList3, returnURL)
								}
							}
						}
					}

					if (initialStatusCode!=getTmpStatusCode4) {
						tmpResultList3 = append(tmpResultList3, v)
					} else {
						if (getTmpStatusCode4==200) {
							if (initialPageSize!=getLenBody4) {
								tmpResultList3 = append(tmpResultList3, v)
							}
						}
					}
				} else {
					for counter1>numberOfa-1 {							
						var uriPath1=""				
						if counter1==numberOfa {
							uriPath1=strings.Replace(u.Path,"/"+tmpSplit2[counter1],"/",1)
						} else {
							uriPath1=strings.Replace(u.Path,"/"+tmpSplit2[counter1],"/xxx",1)
						}
						var newURL=u.Scheme+"://"+u.Host+uriPath1

						req1, err := http.NewRequest("GET", newURL, nil)
						if err==nil {
							req1.Header.Add("User-Agent", userAgent)
							req1.SetBasicAuth(identUser, identPass)
							resp1, err := client.Do(req1)		
							if resp1!=nil{					
								defer resp1.Body.Close()
							}
							if err==nil {								
								body, err := ioutil.ReadAll(resp1.Body)
								if err==nil {
									lenBody = len(body)

									if enableDebug==true {
										fmt.Println("[debug17] "+newURL+" ["+strconv.Itoa(resp1.StatusCode)+"]")
									}
									if resp1.StatusCode==initialStatusCode && initialPageSize==lenBody {
										u1, err := url.Parse(newURL)
										if err != nil {
											panic(err)
										}
										tmpSplit3 :=strings.Split(u1.Path,"/")
										if len(tmpSplit3)>3 {
											var uriPath2=strings.Replace(u1.Path,"/"+tmpSplit3[2]+"/","",1)
											var newURL1=u.Scheme+"://"+u.Host+uriPath2
											req2, err := http.NewRequest("GET", newURL1, nil)
											req2.Header.Add("User-Agent", userAgent)
											req2.SetBasicAuth(identUser, identPass)
											resp2, err := client.Do(req2)														
											if resp2!=nil{					
												defer resp2.Body.Close()
											}
											if err==nil {
												body2, _ := ioutil.ReadAll(resp2.Body)
												if enableDebug==true {
													fmt.Println("[debug18] "+newURL1+" ["+strconv.Itoa(resp2.StatusCode)+"]")
												}
									
												var lenBody2 = len(body2)
												if resp2.StatusCode==initialStatusCode && initialPageSize==lenBody2 {
													if strings.HasSuffix(newURL1, "/") {
														newURL1=newURL1[0:len(newURL1)-1]
													}
													if !stringInSlice(newURL1,tmpResultList3) {
														tmpResultList3 = append(tmpResultList3, newURL1)
													} 
												} else {
													if !stringInSlice(v,tmpResultList3) {
														tmpResultList3 = append(tmpResultList3, v)
													}
												}
											}
										} else {			
											var newURL1 = newURL[0:len(newURL)-1]
											req2, _ := http.NewRequest("GET", newURL1, nil)
											req2.Header.Add("User-Agent", userAgent)
											req2.SetBasicAuth(identUser, identPass)
											resp2, _ := client.Do(req2)
											if resp2!=nil{					
												defer resp2.Body.Close()
											}
											if resp2.StatusCode==resp1.StatusCode {
												newURL=newURL1
											}
											if enableDebug==true {
												fmt.Println("[debug19] "+newURL1+" ["+strconv.Itoa(resp2.StatusCode)+"]")
											}
											if resp1.StatusCode==initialStatusCode && initialPageSize==lenBody {																												
												if !stringInSlice(newURL,tmpResultList3) {
													tmpResultList3 = append(tmpResultList3, newURL)
												}
											}
										}
									} else {
										u1, err := url.Parse(v)
										if err != nil {
											panic(err)
										}
										var newURL2=u1.Scheme+"://"+u1.Host
										req2, _ := http.NewRequest("GET", newURL2, nil)
										req2.Header.Add("User-Agent", userAgent)
										req2.SetBasicAuth(identUser, identPass)
										resp2, err := client.Do(req2)														
										if resp2!=nil{					
											defer resp2.Body.Close()
										}
										if enableDebug==true {
											fmt.Println("[debug20] "+newURL2+" ["+strconv.Itoa(resp2.StatusCode)+"]")
										}
										if resp2.StatusCode==resp1.StatusCode {
											tmpResultList3 = append(tmpResultList3, newURL2)
										} else {
											tmpResultList3 = append(tmpResultList3, v)
										}
									}
								}
							} else {
								fmt.Println(err)
							}
						} else {
							fmt.Println(err)
						}
						counter1-=1
					}
				}
			}
		}
	}	
	RemoveDuplicates(&tmpResultList3)
	sort.Strings(tmpResultList3)

	var finalResults []string

	for _, v := range tmpResultList3 {
		timeout := time.Duration(time.Duration(timeoutSec) * time.Second)
		client := http.Client{
			Timeout: timeout,
			CheckRedirect: redirectPolicyFunc,
		}
		v1, err := url.Parse(v)
    	if err != nil {
    	    panic(err)
    	}		
    	if v1.Path!="/" {
			req2, err := http.NewRequest("GET", v, nil)
			req2.Header.Add("User-Agent", userAgent)
			req2.SetBasicAuth(identUser, identPass)
			resp2, err := client.Do(req2)														
			if resp2!=nil{					
				defer resp2.Body.Close()
			}
			if err==nil {
				if checkStatusCode(resp2.StatusCode)==true {
					body2, err2 := ioutil.ReadAll(resp2.Body)				
					if err2==nil {
						s, err3 := goscraper.Scrape(v, 5)
						if err3==nil {
							var tmpTitle2 = ""
							tmpTitle2=strings.TrimSpace(s.Preview.Title)						
							var lenBody2 = len(body2)
							if !stringInSlice(v,tmpResultList4) {
								u, err := url.Parse(v)
								if err != nil {
									panic(err)
								}
								if len(u.Path)>0 {
									if intelligentMode==true {
										var returnURL=""
										if resp2.StatusCode==401 {
											returnURL=pathPrediction(v,resp2.StatusCode)
											if len(returnURL)>0 {
												if returnURL!=v {
													if returnURL!=u.Scheme+"://"+u.Host {
														if !stringInSlice(returnURL,tmpFoundList) {
															if enableDebug==true {
																fmt.Println("[cleanup1]")
															}
															tmpTitle2=strings.Replace(tmpTitle2,"\n"," ",1)
															fmt.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",returnURL, color.BlueString(strconv.Itoa(resp2.StatusCode)),  lenBody2, tmpTitle2)								
															log.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",returnURL, color.BlueString(strconv.Itoa(resp2.StatusCode)),  lenBody2, tmpTitle2)											
															tmpFoundList=append(tmpFoundList,returnURL)
														}
													}
												} else {
													if enableDebug==true {
														fmt.Println("[cleanup2]")
													}
													tmpTitle2=strings.Replace(tmpTitle2,"\n"," ",1)
													fmt.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)),  lenBody2, tmpTitle2)								
													log.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)),  lenBody2, tmpTitle2)
												}
											}
										} else {
											returnURL=pathPrediction(v,resp2.StatusCode)
											if returnURL!=u.Scheme+"://"+u.Host {
												if len(returnURL)>0 {
													if !stringInSlice(returnURL,tmpFoundList) {
														if enableDebug==true {
															fmt.Println("[cleanup3]")
														}
														tmpTitle2=strings.Replace(tmpTitle2,"\n"," ",1)
														fmt.Printf(color.BlueString("[Found]")+" %s [code:%s] [%d] [%s]\n",returnURL, color.BlueString(strconv.Itoa(resp2.StatusCode)),  lenBody2, tmpTitle2)								
														log.Printf(color.BlueString("[Found]")+" %s [code:%s] [%d] [%s]\n",returnURL, color.BlueString(strconv.Itoa(resp2.StatusCode)),  lenBody2, tmpTitle2)
														tmpFoundList=append(tmpFoundList,returnURL)
														finalResults=append(finalResults,returnURL)
													}
												}
											}
										}
									} else {
										tmpTitle2=strings.Replace(tmpTitle2,"\n"," ",1)
										if enableDebug==true {
											fmt.Println("[cleanup4]")
										}
										fmt.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)),  lenBody2, tmpTitle2)								
										log.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)),  lenBody2, tmpTitle2)
									}
									tmpResultList4 = append(tmpResultList4,v)
								}
							}
						} else { 
							if !stringInSlice(v,tmpResultList4) {
								u, err := url.Parse(v)
								if err != nil {
									panic(err)
								}
								if len(u.Path)>0 {
									if intelligentMode==true {
										var returnURL=""
										if resp2.StatusCode==401 {
											returnURL=pathPrediction(v,resp2.StatusCode)
											if len(returnURL)>0 {
												if returnURL!=v {
													if returnURL!=u.Scheme+"://"+u.Host {
														if !stringInSlice(returnURL,tmpFoundList) {
															fmt.Printf(color.BlueString("[Found]")+" %s [%s]\n",returnURL, color.BlueString(strconv.Itoa(resp2.StatusCode)))								
															log.Printf(color.BlueString("[Found]")+" %s [%s]\n",returnURL, color.BlueString(strconv.Itoa(resp2.StatusCode)))
															tmpFoundList=append(tmpFoundList,returnURL)
														}
													}
												} else {
													fmt.Printf(color.BlueString("[Found]")+" %s [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)))								
													log.Printf(color.BlueString("[Found]")+" %s [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)))
												}
											}
										} else {
											fmt.Printf(color.BlueString("[Found]")+" %s [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)))								
											log.Printf(color.BlueString("[Found]")+" %s [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)))
										}										
									} else {
										fmt.Printf(color.BlueString("[Found]")+" %s [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)))								
										log.Printf(color.BlueString("[Found]")+" %s [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)))
									}
								}
							}
						}
					} else {
						if !stringInSlice(v,tmpResultList4) {
							u, err := url.Parse(v)
							if err != nil {
								panic(err)
							}
							if len(u.Path)>0 {						
								var returnURL=""
								if resp2.StatusCode==401 {
									returnURL=pathPrediction(v,resp2.StatusCode)
									if len(returnURL)>0 {
										if returnURL!=v {
											if returnURL!=u.Scheme+"://"+u.Host {
												fmt.Printf(color.BlueString("[Found]")+" %s [%s] --> %s\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)),returnURL)								
												log.Printf(color.BlueString("[Found]")+" %s [%s] --> %s\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)),returnURL)
											} else {
												fmt.Printf(color.BlueString("[Found]")+" %s [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)))								
												log.Printf(color.BlueString("[Found]")+" %s [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)))
											}
										}
									}
								} else {
									fmt.Printf(color.BlueString("[Found]")+" %s [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)))								
									log.Printf(color.BlueString("[Found]")+" %s [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)))
								}								
								tmpResultList4 = append(tmpResultList4,v)
							}
						}
					}
				}
			}
		}
	if len(finalResults)>0 {
		if lookupMode==true {
			//fmt.Println("\n[*] Looking up URI Paths in ExploitDB/Packetstorm/Metasploit DB")
			fmt.Println("\n[*] Looking up URI Paths in ExploitDB Database")
			fmt.Println("[Source]\t[Filename]\t[URI Path]\t\t[Vuln Category]")
			for _, v := range finalResults {
				u, err := url.Parse(v)
				if err == nil {
					lookupURI(u.Path)
				}	
			}	
		}
	}							
	}}
func checkURL(urlChan chan string) {
	var tmpResultList3 []string
    for v := range urlChan {    
		timeout := time.Duration(time.Duration(timeoutSec) * time.Second)
		client := http.Client{
			Timeout: timeout,
			CheckRedirect: redirectPolicyFunc,
		}
		if proxyMode==true {
			url_i := url.URL{}
			url_proxy, _ := url_i.Parse(proxy_addr)
			http.DefaultTransport.(*http.Transport).Proxy = http.ProxyURL(url_proxy)
		}
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		req, err := http.NewRequest("GET", v, nil)
		if err==nil {
			req.Header.Add("User-Agent", userAgent)
			req.SetBasicAuth(identUser, identPass)
			resp, err := client.Do(req)		
			if resp!=nil{					
				defer resp.Body.Close()
			} else {
				if err==nil {
					s, err := goscraper.Scrape(v, 5)
					var lenBody = 0
					var tmpTitle = ""
					if err==nil {
						tmpTitle=strings.TrimSpace(s.Preview.Title)						
						body, err3 := ioutil.ReadAll(resp.Body)
						if err3==nil {
							lenBody = len(body)
						}
						if checkStatusCode(resp.StatusCode)==true {
							if intelligentMode==false {
								if resp.StatusCode==401 {
									tmpTitle=strings.Replace(tmpTitle,"\n"," ",1)
									fmt.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",v, color.GreenString(strconv.Itoa(resp.StatusCode)),  lenBody, tmpTitle)								
									log.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",v, color.GreenString(strconv.Itoa(resp.StatusCode)),  lenBody, tmpTitle)
								} else if (resp.StatusCode==200) {
									tmpTitle=strings.Replace(tmpTitle,"\n"," ",1)
									fmt.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",v, color.BlueString(strconv.Itoa(resp.StatusCode)),  lenBody, tmpTitle)								
									log.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",v, color.BlueString(strconv.Itoa(resp.StatusCode)),  lenBody, tmpTitle)
								} else {
									tmpTitle=strings.Replace(tmpTitle,"\n"," ",1)
									fmt.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",v, color.RedString(strconv.Itoa(resp.StatusCode)),  lenBody, tmpTitle)								
									log.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",v, color.RedString(strconv.Itoa(resp.StatusCode)),  lenBody, tmpTitle)
								}				
							} else {
								var initialStatusCode=resp.StatusCode
								var initialPageSize=lenBody

								u, err := url.Parse(v)
								if err != nil {
									panic(err)
								}
								numberOfa := strings.Count(u.Path, "/")				
								tmpSplit2 :=strings.Split(u.Path,"/")
								var counter1=numberOfa
								if numberOfa<3 {	
									// var newURL3=u.Scheme+"://"+u.Host
									// getTmpFinalURL3,getTmpTitle3,getTmpStatusCode3,getLenBody3:=getPage(newURL3)
									var getTmpStatusCode4=0
									var getLenBody4=0
									var newURL4=""
					
									if !strings.HasSuffix(v,"/") {
										if strings.Contains( u.Path,".") {	
											tmpSplit1 :=strings.Split(u.Path,".")
											var fileExt = ("."+tmpSplit1[len(tmpSplit1)-1])

											tmpSplit3 :=strings.Split(u.Path,fileExt)
											newURL4=u.Scheme+"://"+u.Host+tmpSplit3[0]+"xxx"+fileExt
											_,_,getTmpStatusCode4,getLenBody4=getPage(newURL4)
											if enableDebug==true {
												fmt.Println("[debug15] "+newURL4+" ["+strconv.Itoa(getTmpStatusCode4)+"]")
											}
											if (resp.StatusCode==getTmpStatusCode4 && (getTmpStatusCode4!=200 && getTmpStatusCode4!=401 && getTmpStatusCode4!=405)) {
												var returnURL=u.Scheme+"://"+u.Host
												tmpResultList3 = append(tmpResultList3, returnURL)
											} else {
												if lenBody!=getLenBody4 {
													var returnURL=v
													tmpResultList3 = append(tmpResultList3, returnURL)
												}
											}
										}
									}

									if (initialStatusCode!=getTmpStatusCode4) {
										tmpResultList3 = append(tmpResultList3, v)
									} else {
										if (getTmpStatusCode4==200) {
											if (initialPageSize!=getLenBody4) {
												tmpResultList3 = append(tmpResultList3, v)
											}
										}
									}
								} else {
									for counter1>numberOfa-1 {							
										var uriPath1=""				
										if counter1==numberOfa {
											uriPath1=strings.Replace(u.Path,"/"+tmpSplit2[counter1],"/",1)
										} else {
											uriPath1=strings.Replace(u.Path,"/"+tmpSplit2[counter1],"/xxx",1)
										}
										var newURL=u.Scheme+"://"+u.Host+uriPath1

										req1, err := http.NewRequest("GET", newURL, nil)
										if err==nil {
											req1.Header.Add("User-Agent", userAgent)
											req1.SetBasicAuth(identUser, identPass)
											resp1, err := client.Do(req1)		
											if resp1!=nil{					
												defer resp1.Body.Close()
											}
											if err==nil {								
												body, err := ioutil.ReadAll(resp1.Body)
												if err==nil {
													lenBody = len(body)

													if enableDebug==true {
														fmt.Println("[debug17] "+newURL+" ["+strconv.Itoa(resp1.StatusCode)+"]")
													}
													if resp1.StatusCode==initialStatusCode && initialPageSize==lenBody {
														u1, err := url.Parse(newURL)
														if err != nil {
															panic(err)
														}
														tmpSplit3 :=strings.Split(u1.Path,"/")
														if len(tmpSplit3)>3 {
															var uriPath2=strings.Replace(u1.Path,"/"+tmpSplit3[2]+"/","",1)
															var newURL1=u.Scheme+"://"+u.Host+uriPath2
															req2, err := http.NewRequest("GET", newURL1, nil)
															req2.Header.Add("User-Agent", userAgent)
															req2.SetBasicAuth(identUser, identPass)
															resp2, err := client.Do(req2)														
															if resp2!=nil{					
																defer resp2.Body.Close()
															}
															if err==nil {
																body2, _ := ioutil.ReadAll(resp2.Body)
																if enableDebug==true {
																	fmt.Println("[debug18] "+newURL1+" ["+strconv.Itoa(resp2.StatusCode)+"]")
																}
										
																var lenBody2 = len(body2)
																if resp2.StatusCode==initialStatusCode && initialPageSize==lenBody2 {
																	if strings.HasSuffix(newURL1, "/") {
																		newURL1=newURL1[0:len(newURL1)-1]
																	}
																	if !stringInSlice(newURL1,tmpResultList3) {
																		tmpResultList3 = append(tmpResultList3, newURL1)
																	} 
																} else {
																	if !stringInSlice(v,tmpResultList3) {
																		tmpResultList3 = append(tmpResultList3, v)
																	}
																}
															}
														} else {			
															var newURL1 = newURL[0:len(newURL)-1]
															req2, _ := http.NewRequest("GET", newURL1, nil)
															req2.Header.Add("User-Agent", userAgent)
															req2.SetBasicAuth(identUser, identPass)
															resp2, _ := client.Do(req2)
															if resp2!=nil{					
																defer resp2.Body.Close()
															}
															if resp2.StatusCode==resp1.StatusCode {
																newURL=newURL1
															}
															if enableDebug==true {
																fmt.Println("[debug19] "+newURL1+" ["+strconv.Itoa(resp2.StatusCode)+"]")
															}
															if resp1.StatusCode==initialStatusCode && initialPageSize==lenBody {																												
																if !stringInSlice(newURL,tmpResultList3) {
																	tmpResultList3 = append(tmpResultList3, newURL)
																}
															}
														}
													} else {
														u1, err := url.Parse(v)
														if err != nil {
															panic(err)
														}
														var newURL2=u1.Scheme+"://"+u1.Host
														req2, err := http.NewRequest("GET", newURL2, nil)
														req2.Header.Add("User-Agent", userAgent)
														req2.SetBasicAuth(identUser, identPass)
														resp2, err := client.Do(req2)														
														if resp2!=nil{					
															defer resp2.Body.Close()
														}
														if enableDebug==true {
															fmt.Println("[debug20] "+newURL2+" ["+strconv.Itoa(resp2.StatusCode)+"]")
														}
														if resp2.StatusCode==resp1.StatusCode {
															tmpResultList3 = append(tmpResultList3, newURL2)
														} else {
															tmpResultList3 = append(tmpResultList3, v)
														}
													}
												}
											} else {
												fmt.Println(err)
											}
										} else {
											fmt.Println(err)
										}
										counter1-=1
									}
								}
							}
						}
					}
				}
			}
		}	
		currentListCount1+=1
    }

	RemoveDuplicates(&tmpResultList3)
	sort.Strings(tmpResultList3)
	for _, v := range tmpResultList3 {
		timeout := time.Duration(time.Duration(timeoutSec) * time.Second)
		client := http.Client{
			Timeout: timeout,
			CheckRedirect: redirectPolicyFunc,
		}
		v1, err := url.Parse(v)
    	if err != nil {
    	    panic(err)
    	}		
    	if v1.Path!="/" {
			req2, err := http.NewRequest("GET", v, nil)
			req2.Header.Add("User-Agent", userAgent)
			req2.SetBasicAuth(identUser, identPass)
			resp2, err := client.Do(req2)														
			if resp2!=nil{					
				defer resp2.Body.Close()
			}
			if err==nil {
				if checkStatusCode(resp2.StatusCode)==true {
					body2, err2 := ioutil.ReadAll(resp2.Body)				
					if err2==nil {
						s, err3 := goscraper.Scrape(v, 5)
						if err3==nil {
							var tmpTitle2 = ""
							tmpTitle2=strings.TrimSpace(s.Preview.Title)						
							var lenBody2 = len(body2)
							if !stringInSlice(v,tmpResultList4) {
								u, err := url.Parse(v)
								if err != nil {
									panic(err)
								}
								if len(u.Path)>0 {
									if intelligentMode==true {
										var returnURL=""
										if resp2.StatusCode==401 {
											returnURL=pathPrediction(v,resp2.StatusCode)
											if len(returnURL)>0 {
												if returnURL!=v {
													if returnURL!=u.Scheme+"://"+u.Host {
														if !stringInSlice(returnURL,tmpFoundList) {
															if enableDebug==true {
																fmt.Println("[cleanup1]")
															}
															tmpTitle2=strings.Replace(tmpTitle2,"\n"," ",1)
															fmt.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",returnURL, color.BlueString(strconv.Itoa(resp2.StatusCode)),  lenBody2, tmpTitle2)								
															log.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",returnURL, color.BlueString(strconv.Itoa(resp2.StatusCode)),  lenBody2, tmpTitle2)											
															tmpFoundList=append(tmpFoundList,returnURL)
														}
													}
												} else {
													if enableDebug==true {
														fmt.Println("[cleanup2]")
													}
													tmpTitle2=strings.Replace(tmpTitle2,"\n"," ",1)
													fmt.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)),  lenBody2, tmpTitle2)								
													log.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)),  lenBody2, tmpTitle2)
												}
											}
										} else {
											returnURL=pathPrediction(v,resp2.StatusCode)
											if returnURL!=u.Scheme+"://"+u.Host {
												if len(returnURL)>0 {
													if !stringInSlice(returnURL,tmpFoundList) {
														if enableDebug==true {
															fmt.Println("[cleanup3]")
														}
														tmpTitle2=strings.Replace(tmpTitle2,"\n"," ",1)
														fmt.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",returnURL, color.BlueString(strconv.Itoa(resp2.StatusCode)),  lenBody2, tmpTitle2)								
														log.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",returnURL, color.BlueString(strconv.Itoa(resp2.StatusCode)),  lenBody2, tmpTitle2)
														tmpFoundList=append(tmpFoundList,returnURL)
													}
												}
											}
										}
									} else {
										tmpTitle2=strings.Replace(tmpTitle2,"\n"," ",1)
										if enableDebug==true {
											fmt.Println("[cleanup4]")
										}
										fmt.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)),  lenBody2, tmpTitle2)								
										log.Printf(color.BlueString("[Found]")+" %s [%s] [%d] [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)),  lenBody2, tmpTitle2)
									}
									tmpResultList4 = append(tmpResultList4,v)
								}
							}
						} else { 
							if !stringInSlice(v,tmpResultList4) {
								u, err := url.Parse(v)
								if err != nil {
									panic(err)
								}
								if len(u.Path)>0 {
									if intelligentMode==true {
										var returnURL=""
										if resp2.StatusCode==401 {
											returnURL=pathPrediction(v,resp2.StatusCode)
											if len(returnURL)>0 {
												if returnURL!=v {
													if returnURL!=u.Scheme+"://"+u.Host {
														if !stringInSlice(returnURL,tmpFoundList) {
															fmt.Printf(color.BlueString("[Found]")+" %s [%s]\n",returnURL, color.BlueString(strconv.Itoa(resp2.StatusCode)))								
															log.Printf(color.BlueString("[Found]")+" %s [%s]\n",returnURL, color.BlueString(strconv.Itoa(resp2.StatusCode)))
															tmpFoundList=append(tmpFoundList,returnURL)
														}
													}
												} else {
													fmt.Printf(color.BlueString("[Found]")+" %s [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)))								
													log.Printf(color.BlueString("[Found]")+" %s [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)))
												}
											}
										} else {
											fmt.Printf(color.BlueString("[Found]")+" %s [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)))								
											log.Printf(color.BlueString("[Found]")+" %s [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)))
										}										
									} else {
										fmt.Printf(color.BlueString("[Found]")+" %s [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)))								
										log.Printf(color.BlueString("[Found]")+" %s [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)))
									}
								}
							}
						}
					} else {
						if !stringInSlice(v,tmpResultList4) {
							u, err := url.Parse(v)
							if err != nil {
								panic(err)
							}
							if len(u.Path)>0 {						
								var returnURL=""
								if resp2.StatusCode==401 {
									returnURL=pathPrediction(v,resp2.StatusCode)
									if len(returnURL)>0 {
										if returnURL!=v {
											if returnURL!=u.Scheme+"://"+u.Host {
												fmt.Printf(color.BlueString("[Found]")+" %s [%s] --> %s\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)),returnURL)								
												log.Printf(color.BlueString("[Found]")+" %s [%s] --> %s\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)),returnURL)
											} else {
												fmt.Printf(color.BlueString("[Found]")+" %s [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)))								
												log.Printf(color.BlueString("[Found]")+" %s [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)))
											}
										}
									}
								} else {
									fmt.Printf(color.BlueString("[Found]")+" %s [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)))								
									log.Printf(color.BlueString("[Found]")+" %s [%s]\n",v, color.BlueString(strconv.Itoa(resp2.StatusCode)))
								}								
								tmpResultList4 = append(tmpResultList4,v)
							}
						}
					}
				}
			}
		}
	}
}

func getPageTitle(u string) (string) {
	var tmpTitle=""
	s, err := goscraper.Scrape(u, 5)
	if err==nil {
		tmpTitle = s.Preview.Title
		tmpTitle = strings.TrimSpace(tmpTitle)
		return tmpTitle
	}
	return tmpTitle

}

func testURL(newUrl string) {
    	var newUrl1 = strings.Split(newUrl," | ")
    	newUrl = newUrl1[0]
    	var currentListCount, _ = strconv.Atoi(newUrl1[1])
		timeout := time.Duration(time.Duration(timeoutSec) * time.Second)
		client := http.Client{
			Timeout: timeout,
			CheckRedirect: redirectPolicyFunc,	
		}
		if proxyMode==true {
			url_i := url.URL{}
			url_proxy, _ := url_i.Parse(proxy_addr)
			http.DefaultTransport.(*http.Transport).Proxy = http.ProxyURL(url_proxy)
		}
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		req, err := http.NewRequest("GET", newUrl, nil)
		if err==nil {
			req.Header.Add("User-Agent", userAgent)
			req.SetBasicAuth(identUser, identPass)
			initialStatusCode := ""
			resp, err := client.Do(req)			
			if resp!=nil{					
				defer resp.Body.Close()
			}
			if err!=nil{									
				if (strings.Contains(err.Error(),"i/o timeout") || strings.Contains(err.Error(),"Client.Timeout exceeded") || strings.Contains(err.Error(),"TLS handshake timeout")) {
					fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Timeout"),currentListCount,totalListCount)						
					log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Timeout"),currentListCount,totalListCount)
				} else if strings.Contains(err.Error(),"connection refused") {
					fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Connection Refused"),currentListCount,totalListCount)									
					log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Connection Refused"),currentListCount,totalListCount)
				} else if strings.Contains(err.Error(),"no such host") {
					fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Unknown Host"),currentListCount,totalListCount)									
					log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Unknown Host"),currentListCount,totalListCount)	
				} else if strings.Contains(err.Error(),"connection reset by peer") {
					fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Connection Reset"),currentListCount,totalListCount)									
					log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Connection Reset"),currentListCount,totalListCount)	
				} else if strings.Contains(err.Error(),"tls: no renegotiation") || strings.Contains(err.Error(),"tls: alert(") {
					fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("TLS Error"),currentListCount,totalListCount)	
					log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("TLS Error"),currentListCount,totalListCount)	
				} else if strings.Contains(err.Error(),"stopped after 10 redirects") {
					fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Max Redirect"),currentListCount,totalListCount)	
					log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Max Redirect"),currentListCount,totalListCount)							
				} else if strings.Contains(err.Error()," EOF]") {
					fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("EOF"),currentListCount,totalListCount)	
					log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("EOF"),currentListCount,totalListCount)													
				} else if strings.Contains(err.Error(),"server gave HTTP response to HTTPS client") {
					fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("302"),currentListCount,totalListCount)	
					log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("302"),currentListCount,totalListCount)																
				} else if strings.Contains(err.Error(),"network is unreachable") {
					fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Unreachable"),currentListCount,totalListCount)	
					log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Unreachable"),currentListCount,totalListCount)																
				} else if strings.Contains(err.Error(),"no route to hosts") {
					fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Unreachable"),currentListCount,totalListCount)	
					log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Unreachable"),currentListCount,totalListCount)																
				} else if strings.Contains(err.Error(),"EOF") {
					fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("EOF"),currentListCount,totalListCount)	
					log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("EOF"),currentListCount,totalListCount)																
				} else if strings.Contains(err.Error(),"tls: handshake failure") {
					fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Handshake Failure"),currentListCount,totalListCount)	
					log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Handshake Failure"),currentListCount,totalListCount)																
				} else {
					fmt.Printf("1 %s [%s] [%d of %d]\n",newUrl, color.RedString(err.Error()))
					log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString(err.Error()))
				}
				currentListCount+=1
			} else {
				initialStatusCode = strconv.Itoa(resp.StatusCode)
				initialTmpTitle := ""
				s, err := goscraper.Scrape(newUrl, 5)
				if err==nil {
					initialTmpTitle = s.Preview.Title
				}
				if verboseMode==true {
					var lenBody = 0
					body, err := ioutil.ReadAll(resp.Body)
					if err==nil {
						//errorFound=true
						lenBody = len(body)
					}
					finalURL := resp.Request.URL.String()
					var tmpTitle = ""
					if finalURL==newUrl {
						tmpTitle=getPageTitle(newUrl)
					}		
					if intelligentMode==true && CMSmode==false{
						tmpStatusCode := strconv.Itoa(resp.StatusCode)
						var tmpFound=false
						for _, each := range tmpTitleList { 
							var originalURL=""
							if strings.HasSuffix(each[0],"/") {
								originalURL=each[0]
							} else {
								originalURL=each[0]+"/"
							}
							if strings.Contains(finalURL,originalURL) {
								if newUrl==finalURL { 		
									tmpFound=true			
									if (strings.TrimSpace(each[1])!=strings.TrimSpace(tmpTitle) || len(tmpTitle)<1) {
										if tmpTitle!="Error" && tmpTitle!="Request Rejected" && tmpTitle!="Runtime Error"{
											if checkStatusCode(resp.StatusCode)==true {
												if (each[2]!=strconv.Itoa(lenBody)) {
													if CMSmode==false {
														if each[3]!=initialStatusCode && each[2]!=strconv.Itoa(lenBody){
															var a = [][]string{{newUrl, initialStatusCode, strconv.Itoa(lenBody),initialTmpTitle}}
															tmpResultList = append(tmpResultList,a...)
														}
													}
												} 												
											}
										}  
									} else {
										if (strings.TrimSpace(each[1])==strings.TrimSpace(tmpTitle)) {
											if initialStatusCode!=each[3] {
												var a = [][]string{{newUrl, initialStatusCode, strconv.Itoa(lenBody),initialTmpTitle}}
												tmpResultList = append(tmpResultList,a...)
											}
										}
									}
								}
								if tmpFound==true {
									tmpTitle=strings.Replace(tmpTitle,"\n"," ",1)
									if tmpStatusCode=="200"{
										i, _ :=strconv.Atoi(initialStatusCode)
										if (Excludecode==0 && Excludecode!=i) || (Statuscode!=0 && Statuscode==i) {																				
											fmt.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.BlueString(initialStatusCode),  lenBody, tmpTitle,currentListCount,totalListCount)
											log.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.BlueString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)
										}
									} else if tmpStatusCode=="401"{
										i, _ :=strconv.Atoi(initialStatusCode)
										if (Excludecode==0 && Excludecode!=i) || (Statuscode!=0 && Statuscode==i) {										
											fmt.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.GreenString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)										
											log.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.GreenString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)
										}
									} else {
										if initialStatusCode=="0" {
											fmt.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.RedString(""),  lenBody, tmpTitle, currentListCount,totalListCount)
											log.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.RedString(""),  lenBody, tmpTitle, currentListCount,totalListCount)
										} else {
											i, _ :=strconv.Atoi(initialStatusCode)
											if (Excludecode==0 && Excludecode!=i) || (Statuscode!=0 && Statuscode==i) {										
												fmt.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.RedString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)
												log.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.RedString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)
											}
										}
									}
								}
							}
						}
						if tmpFound==false {
							u, err := url.Parse(newUrl)
							if err != nil {
								panic(err)
							}				
							var newURL2=u.Scheme+"://"+u.Host				
							if resp.StatusCode==401 && initialStatusCode=="401" {
								fmt.Printf("%s [code:%s] [%d of %d]\n",newURL2, color.RedString(initialStatusCode), currentListCount,totalListCount)					
								log.Printf("%s [code:%s] [%d of %d]\n",newURL2, color.RedString(initialStatusCode), currentListCount,totalListCount)
								var a = [][]string{{newURL2, initialStatusCode, "",""}}
								tmpResultList = append(tmpResultList,a...)
							} else if (resp.StatusCode!=401 && initialStatusCode=="401") {
								i, _ :=strconv.Atoi(initialStatusCode)
								if (Excludecode==0 && Excludecode!=i) || (Statuscode!=0 && Statuscode==i) {								
									fmt.Printf("%s [code:%s] [%d of %d]\n",newURL2, color.RedString(initialStatusCode), currentListCount,totalListCount)					
									log.Printf("%s [code:%s] [%d of %d]\n",newURL2, color.RedString(initialStatusCode), currentListCount,totalListCount)
								}
								var a = [][]string{{newURL2, initialStatusCode, "",""}}
								tmpResultList = append(tmpResultList,a...)
							} else {
								if tmpStatusCode=="200"{
									i, _ :=strconv.Atoi(initialStatusCode)
									if (Excludecode==0 && Excludecode!=i) || (Statuscode!=0 && Statuscode==i) {		
										fmt.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.BlueString(initialStatusCode),  lenBody, tmpTitle,currentListCount,totalListCount)
										log.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.BlueString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)
									}
								} else if tmpStatusCode=="401"{
									i, _ :=strconv.Atoi(initialStatusCode)
									if (Excludecode==0 && Excludecode!=i) || (Statuscode!=0 && Statuscode==i) {		
										fmt.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.GreenString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)										
										log.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.GreenString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)
									}
								} else {
									i, _ :=strconv.Atoi(initialStatusCode)
									if (Excludecode==0 && Excludecode!=i) || (Statuscode!=0 && Statuscode==i) {								
										fmt.Printf("3%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.RedString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)
										log.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.RedString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)
									}
								}
							}
						}
					} else {
						tmpStatusCode := strconv.Itoa(resp.StatusCode)
						if Statuscode!=0 {
							if resp.StatusCode==Statuscode {
								i, _ :=strconv.Atoi(initialStatusCode)
								if (Excludecode==0 || Excludecode!=i) && (Statuscode==0 || Statuscode==i) {									
									fmt.Printf("2%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.RedString(tmpStatusCode), lenBody, tmpTitle, currentListCount,totalListCount)					
									log.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.RedString(tmpStatusCode), lenBody, tmpTitle, currentListCount,totalListCount)
								}
								var a = [][]string{{newUrl, tmpStatusCode, strconv.Itoa(lenBody),tmpTitle}}
								tmpResultList = append(tmpResultList,a...)
							} else {
								i, _ :=strconv.Atoi(initialStatusCode)
								if (Excludecode==0 || Excludecode!=i) && (Statuscode==0 || Statuscode==i) {		
									fmt.Printf("1%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.BlueString(initialStatusCode),  lenBody, tmpTitle,currentListCount,totalListCount)
									log.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.BlueString(initialStatusCode),  lenBody, tmpTitle,currentListCount,totalListCount)
								}
							}
						} else {				
							if tmpStatusCode=="200"{
								i, _ :=strconv.Atoi(initialStatusCode)
								if (Excludecode==0 || Excludecode!=i) && (Statuscode==0 || Statuscode==i) {		
									fmt.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.BlueString(tmpStatusCode), lenBody, tmpTitle,currentListCount,totalListCount)					
									log.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.BlueString(tmpStatusCode), lenBody, tmpTitle,currentListCount,totalListCount)
								}
								var a = [][]string{{newUrl, tmpStatusCode, strconv.Itoa(lenBody),tmpTitle}}
								tmpResultList = append(tmpResultList,a...)
							} else if tmpStatusCode=="401"{
								i, _ :=strconv.Atoi(initialStatusCode)
								if (Excludecode==0 || Excludecode!=i) && (Statuscode==0 || Statuscode==i) {							
									fmt.Printf("%s [code:%s]\n",newUrl, color.GreenString(tmpStatusCode))
									log.Printf("%s [code:%s]\n",newUrl, color.GreenString(tmpStatusCode))
								}
								var a = [][]string{{newUrl, tmpStatusCode, "",""}}
								tmpResultList = append(tmpResultList,a...)
							} else {
								i, _ :=strconv.Atoi(initialStatusCode)
								if (Excludecode==0 || Excludecode!=i) && (Statuscode==0 || Statuscode==i) {			
									fmt.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.RedString(tmpStatusCode), lenBody, tmpTitle, currentListCount,totalListCount)	
									log.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.RedString(tmpStatusCode), lenBody, tmpTitle, currentListCount,totalListCount)				
								}
							}
						}
					}
				} else {
					if Statuscode!=0 {
						tmpStatusCode := strconv.Itoa(resp.StatusCode)	
						if resp.StatusCode==Statuscode {	
							fmt.Printf("%s [code:%s]\n",newUrl, color.BlueString(tmpStatusCode))
							log.Printf("%s [code:%s]\n",newUrl, color.BlueString(tmpStatusCode))
							finalURL := resp.Request.URL.String()
							if strings.HasSuffix(finalURL,"/") {
								finalURL=finalURL[0:len(finalURL)-1]
							}
							if finalURL==newUrl {
								if resp.StatusCode!=403 {
									var a = [][]string{{newUrl, tmpStatusCode, "",""}}
									tmpResultList = append(tmpResultList,a...)
								}
							} 
						}

					} else {
						tmpStatusCode := strconv.Itoa(resp.StatusCode)	
						if resp.StatusCode==200 {		
							fmt.Printf("%s [%s]\n",newUrl, color.BlueString(tmpStatusCode))
							log.Printf("%s [%s]\n",newUrl, color.BlueString(tmpStatusCode))
							finalURL := resp.Request.URL.String()
							if strings.HasSuffix(finalURL,"/") {
								finalURL=finalURL[0:len(finalURL)-1]
							}
							if finalURL==newUrl {
								if resp.StatusCode!=403 {
									var a = [][]string{{newUrl, tmpStatusCode, "",""}}
									tmpResultList = append(tmpResultList,a...)
								}
							}
						} else {
							fmt.Printf("%s [%s]\n",newUrl, color.RedString(tmpStatusCode))
							log.Printf("%s [%s]\n",newUrl, color.RedString(tmpStatusCode))
						}				
					}
				}
				resp.Body.Close()
			} 			

		}
		if currentListCount>=totalListCount {
			addToCompleteList(newUrl)	
			reachedTheEnd=true
		} else {
			addToCompleteList(newUrl)
		}
}

func getUrlWorker(urlChan chan string) {
    for newUrl := range urlChan {
    	var newUrl1 = strings.Split(newUrl," | ")
    	newUrl = newUrl1[0]
    	var currentListCount, _ = strconv.Atoi(newUrl1[1])
		timeout := time.Duration(time.Duration(timeoutSec) * time.Second)
		client := http.Client{
			Timeout: timeout,
			CheckRedirect: redirectPolicyFunc,
		}
		if proxyMode==true {
			url_i := url.URL{}
			url_proxy, _ := url_i.Parse(proxy_addr)
			http.DefaultTransport.(*http.Transport).Proxy = http.ProxyURL(url_proxy)
		}
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		req, err := http.NewRequest("GET", newUrl, nil)
		if err==nil {
			req.Header.Add("User-Agent", userAgent)
			req.SetBasicAuth(identUser, identPass)
			initialStatusCode := ""
			resp, err := client.Do(req)			
			if resp!=nil{					
				defer resp.Body.Close()
			}
			if err!=nil{									
				if (strings.Contains(err.Error(),"i/o timeout") || strings.Contains(err.Error(),"Client.Timeout exceeded") || strings.Contains(err.Error(),"TLS handshake timeout")) {
					fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Timeout"),currentListCount,totalListCount)						
					log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Timeout"),currentListCount,totalListCount)
				} else if strings.Contains(err.Error(),"connection refused") {
					fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Connection Refused"),currentListCount,totalListCount)									
					log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Connection Refused"),currentListCount,totalListCount)
				} else if strings.Contains(err.Error(),"no such host") {
					fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Unknown Host"),currentListCount,totalListCount)									
					log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Unknown Host"),currentListCount,totalListCount)	
				} else if strings.Contains(err.Error(),"connection reset by peer") {
					fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Connection Reset"),currentListCount,totalListCount)									
					log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Connection Reset"),currentListCount,totalListCount)	
				} else if strings.Contains(err.Error(),"tls: no renegotiation") {
					fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("TLS Error"),currentListCount,totalListCount)	
					log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("TLS Error"),currentListCount,totalListCount)	
				} else if strings.Contains(err.Error(),"stopped after 10 redirects") {
					fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Max Redirect"),currentListCount,totalListCount)	
					log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Max Redirect"),currentListCount,totalListCount)							
				} else if strings.Contains(err.Error()," EOF]") {
					fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("EOF"),currentListCount,totalListCount)	
					log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("EOF"),currentListCount,totalListCount)													
				} else if strings.Contains(err.Error(),"server gave HTTP response to HTTPS client") {
					fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("302"),currentListCount,totalListCount)	
					log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("302"),currentListCount,totalListCount)																
				} else if strings.Contains(err.Error(),"network is unreachable") {
					fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Unreachable"),currentListCount,totalListCount)	
					log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Unreachable"),currentListCount,totalListCount)																
				} else if strings.Contains(err.Error(),"no route to hosts") {
					fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Unreachable"),currentListCount,totalListCount)	
					log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Unreachable"),currentListCount,totalListCount)																
				} else if strings.Contains(err.Error(),"EOF") {
					fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("EOF"),currentListCount,totalListCount)	
					log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("EOF"),currentListCount,totalListCount)																
				} else if strings.Contains(err.Error(),"tls: handshake failure") {
					fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Handshake Failure"),currentListCount,totalListCount)	
					log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString("Handshake Failure"),currentListCount,totalListCount)																
				} else {
					fmt.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString(err.Error()))
					log.Printf("%s [%s] [%d of %d]\n",newUrl, color.RedString(err.Error()))
				}
				currentListCount+=1
			} else {
				initialStatusCode = strconv.Itoa(resp.StatusCode)
				initialTmpTitle := ""
				s, err := goscraper.Scrape(newUrl, 5)
				if err==nil {
					initialTmpTitle = s.Preview.Title
				}
				if verboseMode==true {
					var lenBody = 0
					body, err := ioutil.ReadAll(resp.Body)
					if err==nil {
						//errorFound=true
						lenBody = len(body)
					}
					finalURL := resp.Request.URL.String()
					var tmpTitle = ""
					if finalURL==newUrl {
						tmpTitle=getPageTitle(newUrl)
					}		
					if intelligentMode==true && CMSmode==false{
						tmpStatusCode := strconv.Itoa(resp.StatusCode)
						var tmpFound=false
						for _, each := range tmpTitleList { 
							var originalURL=""
							if strings.HasSuffix(each[0],"/") {
								originalURL=each[0]
							} else {
								originalURL=each[0]+"/"
							}
							if strings.Contains(finalURL,originalURL) {
								if newUrl==finalURL { 		
									tmpFound=true			
									if (strings.TrimSpace(each[1])!=strings.TrimSpace(tmpTitle) || len(tmpTitle)<1) {
										if tmpTitle!="Error" && tmpTitle!="Request Rejected" && tmpTitle!="Runtime Error"{
											if checkStatusCode(resp.StatusCode)==true {
												if (each[2]!=strconv.Itoa(lenBody)) {
													if CMSmode==false {
														if each[3]!=initialStatusCode && each[2]!=strconv.Itoa(lenBody){
															var a = [][]string{{newUrl, initialStatusCode, strconv.Itoa(lenBody),initialTmpTitle}}
															tmpResultList = append(tmpResultList,a...)
														}
													}
												} 												
											}
										}  
									} else {
										if (strings.TrimSpace(each[1])==strings.TrimSpace(tmpTitle)) {
											if initialStatusCode!=each[3] {
												var a = [][]string{{newUrl, initialStatusCode, strconv.Itoa(lenBody),initialTmpTitle}}
												tmpResultList = append(tmpResultList,a...)
											}
										}
									}
								}
								if tmpFound==true {
									tmpTitle=strings.Replace(tmpTitle,"\n"," ",1)
									if tmpStatusCode=="200"{
										i, _ :=strconv.Atoi(initialStatusCode)
										if (Excludecode==0 || Excludecode!=i) && (Statuscode==0 || Statuscode==i) {																				
											fmt.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.BlueString(initialStatusCode),  lenBody, tmpTitle,currentListCount,totalListCount)
											log.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.BlueString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)
										}
									} else if tmpStatusCode=="401"{
										i, _ :=strconv.Atoi(initialStatusCode)
										if (Excludecode==0 || Excludecode!=i) && (Statuscode==0 || Statuscode==i) {																				
											fmt.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.GreenString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)										
											log.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.GreenString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)
										}
									} else {
										if initialStatusCode=="0" {
											fmt.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.RedString(""),  lenBody, tmpTitle, currentListCount,totalListCount)
											log.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.RedString(""),  lenBody, tmpTitle, currentListCount,totalListCount)
										} else {
											i, _ :=strconv.Atoi(initialStatusCode)
											if (Excludecode==0 || Excludecode!=i) && (Statuscode==0 || Statuscode==i) {		
												fmt.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.RedString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)
												log.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.RedString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)
											}
										}
									}
								}
							}
						}
						if tmpFound==false {
							u, err := url.Parse(newUrl)
							if err != nil {
								panic(err)
							}				
							var newURL2=u.Scheme+"://"+u.Host				
							if resp.StatusCode==401 && initialStatusCode=="401" {
								i, _ :=strconv.Atoi(initialStatusCode)
								if (Excludecode==0 || Excludecode!=i) && (Statuscode==0 || Statuscode==i) {																		
									fmt.Printf("%s [code:%s] [%d of %d]\n",newURL2, color.RedString(initialStatusCode), currentListCount,totalListCount)					
									log.Printf("%s [code:%s] [%d of %d]\n",newURL2, color.RedString(initialStatusCode), currentListCount,totalListCount)
								}
								var a = [][]string{{newURL2, initialStatusCode, "",""}}
								tmpResultList = append(tmpResultList,a...)
							} else if (resp.StatusCode!=401 && initialStatusCode=="401") {
								i, _ :=strconv.Atoi(initialStatusCode)
								if (Excludecode==0 || Excludecode!=i) && (Statuscode==0 || Statuscode==i) {										
									fmt.Printf("%s [code:%s] [%d of %d]\n",newURL2, color.RedString(initialStatusCode), currentListCount,totalListCount)					
									log.Printf("%s [code:%s] [%d of %d]\n",newURL2, color.RedString(initialStatusCode), currentListCount,totalListCount)
								}
								var a = [][]string{{newURL2, initialStatusCode, "",""}}
								tmpResultList = append(tmpResultList,a...)
							} else {
								if tmpStatusCode=="200"{
									i, _ :=strconv.Atoi(initialStatusCode)
									if (Excludecode==0 || Excludecode!=i) && (Statuscode==0 || Statuscode==i) {																		
										fmt.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.BlueString(initialStatusCode),  lenBody, tmpTitle,currentListCount,totalListCount)
										log.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.BlueString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)
									}
								} else if tmpStatusCode=="401"{
									i, _ :=strconv.Atoi(initialStatusCode)
									if (Excludecode==0 || Excludecode!=i) && (Statuscode==0 || Statuscode==i) {											
										fmt.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.GreenString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)										
										log.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.GreenString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)
									}
								} else {
									i, _ :=strconv.Atoi(initialStatusCode)
									if (Excludecode==0 || Excludecode!=i) && (Statuscode==0 || Statuscode==i) {			
										fmt.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.RedString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)
										log.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.RedString(initialStatusCode),  lenBody, tmpTitle, currentListCount,totalListCount)
									}
								}
							}
						}
					} else {
						tmpStatusCode := strconv.Itoa(resp.StatusCode)
						if Statuscode!=0 {
							if resp.StatusCode==Statuscode {
								i, _ :=strconv.Atoi(initialStatusCode)
								if (Excludecode==0 || Excludecode!=i) && (Statuscode==0 || Statuscode==i) {																	
									fmt.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.RedString(tmpStatusCode), lenBody, tmpTitle, currentListCount,totalListCount)					
									log.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.RedString(tmpStatusCode), lenBody, tmpTitle, currentListCount,totalListCount)
								}
								var a = [][]string{{newUrl, tmpStatusCode, strconv.Itoa(lenBody),tmpTitle}}
								tmpResultList = append(tmpResultList,a...)
							} else {
								i, _ :=strconv.Atoi(initialStatusCode)
								if (Excludecode==0 || Excludecode!=i) && (Statuscode==0 || Statuscode==i) {		
									fmt.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.BlueString(initialStatusCode),  lenBody, tmpTitle,currentListCount,totalListCount)
									log.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.BlueString(initialStatusCode),  lenBody, tmpTitle,currentListCount,totalListCount)
								}
							}
						} else {				
							if tmpStatusCode=="200"{
								i, _ :=strconv.Atoi(initialStatusCode)
								if (Excludecode==0 || Excludecode!=i) && (Statuscode==0 || Statuscode==i) {											
									fmt.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.BlueString(tmpStatusCode), lenBody, tmpTitle,currentListCount,totalListCount)					
									log.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.BlueString(tmpStatusCode), lenBody, tmpTitle,currentListCount,totalListCount)
								}
								var a = [][]string{{newUrl, tmpStatusCode, strconv.Itoa(lenBody),tmpTitle}}
								tmpResultList = append(tmpResultList,a...)
							} else if tmpStatusCode=="401"{
								i, _ :=strconv.Atoi(initialStatusCode)
								if (Excludecode==0 || Excludecode!=i) && (Statuscode==0 || Statuscode==i) {											
									fmt.Printf("%s [code:%s]\n",newUrl, color.GreenString(tmpStatusCode))
									log.Printf("%s [code:%s]\n",newUrl, color.GreenString(tmpStatusCode))
								}
								var a = [][]string{{newUrl, tmpStatusCode, "",""}}
								tmpResultList = append(tmpResultList,a...)
							} else {
								i, _ :=strconv.Atoi(initialStatusCode)
								if (Excludecode==0 || Excludecode!=i) && (Statuscode==0 || Statuscode==i) {		
									fmt.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.RedString(tmpStatusCode), lenBody, tmpTitle, currentListCount,totalListCount)	
									log.Printf("%s [code:%s] [%d] [%s] [%d of %d]\n",newUrl, color.RedString(tmpStatusCode), lenBody, tmpTitle, currentListCount,totalListCount)				
								}
							}
						}
					}
				} else {
					if Statuscode!=0 {
						tmpStatusCode := strconv.Itoa(resp.StatusCode)	
						if resp.StatusCode==Statuscode {	
							i, _ :=strconv.Atoi(initialStatusCode)
							if (Excludecode==0 || Excludecode!=i) && (Statuscode==0 || Statuscode==i) {																	
								fmt.Printf("%s [code:%s]\n",newUrl, color.BlueString(tmpStatusCode))
								log.Printf("%s [code:%s]\n",newUrl, color.BlueString(tmpStatusCode))
							}
							finalURL := resp.Request.URL.String()
							if strings.HasSuffix(finalURL,"/") {
								finalURL=finalURL[0:len(finalURL)-1]
							}
							if finalURL==newUrl {
								if resp.StatusCode!=403 {
									var a = [][]string{{newUrl, tmpStatusCode, "",""}}
									tmpResultList = append(tmpResultList,a...)
								}
							} 
						}

					} else {
						tmpStatusCode := strconv.Itoa(resp.StatusCode)	
						if resp.StatusCode==200 {		
							i, _ :=strconv.Atoi(initialStatusCode)
							if (Excludecode==0 || Excludecode!=i) && (Statuscode==0 || Statuscode==i) {											
								fmt.Printf("%s [code:%s]\n",newUrl, color.BlueString(tmpStatusCode))
								log.Printf("%s [code:%s]\n",newUrl, color.BlueString(tmpStatusCode))
							}
							finalURL := resp.Request.URL.String()
							if strings.HasSuffix(finalURL,"/") {
								finalURL=finalURL[0:len(finalURL)-1]
							}
							if finalURL==newUrl {
								if resp.StatusCode!=403 {
									var a = [][]string{{newUrl, tmpStatusCode, "",""}}
									tmpResultList = append(tmpResultList,a...)
								}
							}
						} else {
							i, _ :=strconv.Atoi(initialStatusCode)
							if (Excludecode==0 || Excludecode!=i) && (Statuscode==0 || Statuscode==i) {		
								fmt.Printf("%s [code:%s]\n",newUrl, color.RedString(tmpStatusCode))
								log.Printf("%s [code:%s]\n",newUrl, color.RedString(tmpStatusCode))
							}
						}
					}
				}
				resp.Body.Close()
			} 			

		}
		if currentListCount>=totalListCount {
			addToCompleteList(newUrl)					
			reachedTheEnd=true
		} else {
			addToCompleteList(newUrl)
		}

    }
}

func readLines(path string) ([]string, error) {
  file, err := os.Open(path)
  if err != nil {
    return nil, err
  }
  defer file.Close()

  var lines []string
  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
    lines = append(lines, scanner.Text())
  }
  return lines, scanner.Err()
}

func BytesToString(data []byte) string {
	return string(data[:])
}

func stringInSlice(str string, list []string) bool {
 	for _, v := range list {
 		if v == str {
 			return true
 		}
 	}
 	return false
}
 


func RemoveDuplicates(xs *[]string) {
	found := make(map[string]bool)
	j := 0
	for i, x := range *xs {
		if !found[x] {
			found[x] = true
			(*xs)[j] = (*xs)[i]
			j++
		}
	}
	*xs = (*xs)[:j]
}

func DownloadFile(filepath string, url string) error {
    out, err := os.Create(filepath)
    if err != nil {
        return err
    }
    defer out.Close()
    resp, err := http.Get(url)
    if err != nil {
    	fmt.Println(err)
        return err
    }
    defer resp.Body.Close()
    _, err = io.Copy(out, resp.Body)
    if err != nil {
    	fmt.Println(err)
        return err
    }
    return nil
}

type argT struct {
	cli.Helper
	Filename string `cli:"U,filename" usage:"File containing list of websites"`
	URLpath string `cli:"u,url" usage:"Url of website"`
	PFilename string `cli:"P,Paths" usage:"File containing list of URI paths"`
	Path string `cli:"p,path" usage:"URI path"`
	Pathsource string `cli:"s,source" usage:"Path source (default | msf | exploitdb | exploitdb-asp | exploitdb-aspx | exploitdb-cfm | exploitdb-cgi | exploitdb-cfm | exploitdb-jsp | exploitdb-perl | exploitdb-php | exploitdb-others | RobotsDisallowed | SecLists)"`
	Threads int  `cli:"n,threads" usage:"No of concurrent threads (default: 2)"`
	Statuscode int  `cli:"c" usage:"Show only certain status code (e.g. -c 200)"`
	Excludecode int  `cli:"e" usage:"Exclude certain status code (e.g. -e 404)"`	
	Intellimode bool `cli:"i" usage:"Intelligent mode"`
	Verbose bool `cli:"v,verbose" usage:"Verbose mode"`
	CMSmode bool `cli:"cms" usage:"Fingerprint CMS"`
	SpreadMode bool `cli:"x" usage:"Test a URI path across all target hosts instead of testing all URI paths against a host before moving onto next host"`
	Logfilename string `cli:"l,log" usage:"Output to log file"`
	ContinueNum int  `cli:"r" usage:"Resume from x as in [x of 9999]"`	
	Proxyhost string `cli:"pHost" usage:"IP of HTTP proxy"`
	Proxyport string `cli:"pPort" usage:"Port of HTTP proxy (default 8080)"`
	Uagent string `cli:"ua" usage:"Set User-Agent"`
	Timeoutsec int `cli:"timeout" usage:"Set timeout to x seconds"`
	Updatemode bool `cli:"update" usage:"Update URI path wordlists from Github"`
	Skipmode bool `cli:"skip" usage:"Skip sites that don't give any useful results (e.g. OWA, VPN, etc)"`
	Confirmmode bool `cli:"confirm" usage:"Confirm using more than 100 threads (use with -n option)"`
	Lookupmode bool `cli:"q,query" usage:"Lookup URI paths that were found against ExploitDB)"`
	Credentials string `cli:"d,ident" usage:"Set basicAuth user:pass"`
}

func main() {
	wpFileList	   = append(wpFileList,"/readme.html")
	joomlaFileList = append(joomlaFileList,"/administrator/manifests/files/joomla.xml")
	joomlaFileList = append(joomlaFileList,"/administrator/language/en-GB/en-GB.xml")
	drupalFileList = append(drupalFileList,"/CHANGELOG.txt")
	
	filename1 := ""
	pFilename := ""
	uriPath := ""
	
	whitelistList = append(whitelistList, "Outlook Web App")
	whitelistList = append(whitelistList, "Netscaler Gateway")
	whitelistList = append(whitelistList, "GlobalProtect Portal")
		
	var contentList []string
	var pathList []string
	
	cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)
		if argv.Excludecode>0 {
			Excludecode=argv.Excludecode
		}
		if argv.Lookupmode {
			lookupMode = true
		}			
		if argv.Timeoutsec>0 {
			timeoutSec = argv.Timeoutsec
		}
		if len(argv.Uagent)>0 {
			userAgent=argv.Uagent
		}		
		if len(argv.Proxyhost)>0 {
			if len(argv.Proxyport)>0 {
				proxy_addr="http://"+argv.Proxyhost+":"+argv.Proxyport
			} else {
				proxy_addr="http://"+argv.Proxyhost+":8080"
			}
			proxyMode=true
		}
		if argv.ContinueNum>0 {
			ContinueNum = argv.ContinueNum
		}
		if len(argv.Logfilename)>0 {
			logfileF, err := os.OpenFile(argv.Logfilename, os.O_WRONLY|os.O_CREATE|os.O_APPEND,0644)
			if err != nil {
					log.Fatal(err)
			}   
			defer logfileF.Close()
			log.SetOutput(logfileF)
		} else {
			logfileF, err := os.OpenFile("tmp.log", os.O_WRONLY|os.O_CREATE,0644)
			if err != nil {
					log.Fatal(err)
			}   
			defer logfileF.Close()
			log.SetOutput(logfileF)
		}
		if len(argv.Credentials)>0 {
			s := strings.Split(argv.Credentials, ":")
			identUser, identPass = s[0], s[1]
		}
		filename1 = argv.Filename
		pFilename = argv.PFilename
		Pathsource = argv.Pathsource

		if argv.SpreadMode {
			SpreadMode = true
		}
		if argv.Statuscode>0 {
			Statuscode = argv.Statuscode
		}
		if argv.Intellimode {
			intelligentMode = true
		}
		if argv.Verbose {
			verboseMode = true
		}		
		if len(argv.Path)>0 { 
			uriPath = argv.Path
		}
		if argv.Threads>0 {
			if argv.Threads>99 {
				if !argv.Confirmmode {
					fmt.Println("[-] Please use the --confirm option if you want to use more than 100 threads.")
					os.Exit(3)
				}
			}
			workersCount = argv.Threads			
		}
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			sig := <-sigs
			fmt.Println(sig)
			cleanup()
			os.Exit(3)
		}()		
		
		if len(pFilename)>0 {		
			_, err1 := os.Stat(pFilename)
			if os.IsNotExist(err1) {
				fmt.Printf("[*] File %s not exists\n", pFilename)
				os.Exit(3)
			}
			lines, err2 := readLines(pFilename)
			if err2==nil {
				for _, v := range lines {
					v=strings.TrimSpace(v)
					if len(v)>0 {
						pathList = append(pathList, v)
					}
				}
			}
		} 
		if len(uriPath)>0 {
			pathList = append(pathList, uriPath)
		}
		if len(Pathsource)>0 { 
			if Pathsource!="default" && Pathsource!="msf" && Pathsource!="exploitdb" && Pathsource!="exploitdb-asp" && Pathsource!="exploitdb-aspx" && Pathsource!="exploitdb-cfm" && Pathsource!="exploitdb-cgi" && Pathsource!="exploitdb-cfm" && Pathsource!="exploitdb-jsp" && Pathsource!="exploitdb-perl" && Pathsource!="exploitdb-php" && Pathsource!="RobotsDisallowed" && Pathsource!="SecLists" {
				fmt.Println("[*] Please select a valid Path source")
				os.Exit(3)
			}
		}
		if Pathsource=="default" {
			pFilename = "defaultPaths.txt"
			_, err1 := os.Stat("defaultPaths.txt")
			if os.IsNotExist(err1) {
				fileUrl := "https://raw.githubusercontent.com/milo2012/pathbrute/master/defaultPaths.txt"
				fmt.Println("[+] Downloading: "+fileUrl)
				err := DownloadFile("defaultPaths.txt", fileUrl)
				if err!=nil {
						if strings.Contains(err.Error(),"no such host") {
							fmt.Println("[*] Error: ",err)
							os.Exit(3)
						} else {
							fmt.Println("[*] Error: ",err)
						}						
				}
			}
			lines, err2 := readLines(pFilename)
			if err2==nil {
				for _, v := range lines {
						v=strings.TrimSpace(v)
						if len(v)>0 {
							pathList = append(pathList, v)
						}
					}		
			}
		}		
		if Pathsource=="msf" {
			fileUrl := "https://raw.githubusercontent.com/milo2012/metasploitHelper/master/pathList.txt"
			tokens := strings.Split(fileUrl,"/")
			var extractFilename=tokens[len(tokens)-1]
			if argv.Updatemode {
				statusCheck:=checkAndUpdate(fileUrl)
				if statusCheck==false {
					fmt.Println("[+] No update required for "+extractFilename)
				} else {
					fmt.Println("[+] Latest version of "+extractFilename+" has been downloaded")
				}
				lines, err2 := readLines(extractFilename)
				if err2==nil {
					for _, v := range lines {
						v=strings.TrimSpace(v)
						if len(v)>0 {
							pathList = append(pathList, v)
						}
					}		
				} else {
					fmt.Println(err2)
				}
			} else {
				_, err1 := os.Stat(extractFilename)
				if os.IsNotExist(err1) {
					fmt.Println("[+] Downloading: "+fileUrl)
					err := DownloadFile(extractFilename, fileUrl)
					if err!=nil {
						if strings.Contains(err.Error(),"no such host") {
							fmt.Println("[*] Error: ",err)
							os.Exit(3)
						} else {
							fmt.Println("[*] Error: ",err)
						}						
					}
				}
				lines, err2 := readLines(extractFilename)
				if err2==nil {
					for _, v := range lines {
						v=strings.TrimSpace(v)
						if len(v)>0 {
							pathList = append(pathList, v)
						}
					}		
				} else {
					fmt.Println(err2)
				}

			}
		}
		if Pathsource=="exploitdb" {
			fileUrl := "https://raw.githubusercontent.com/milo2012/pathbrute/master/exploitdb_all.txt"
			tokens := strings.Split(fileUrl,"/")
			var extractFilename=tokens[len(tokens)-1]
			if argv.Updatemode {
				statusCheck:=checkAndUpdate(fileUrl)
				if statusCheck==false {
					fmt.Println("[+] No update required for "+extractFilename)
				} else {
					fmt.Println("[+] Latest version of "+extractFilename+" has been downloaded")
				}
				lines, err2 := readLines(extractFilename)
				if err2==nil {
					for _, v := range lines {
						v=strings.TrimSpace(v)
						if len(v)>0 {
							pathList = append(pathList, v)
						}
					}		
				} else {
					fmt.Println(err2)
				}
			} else {
				_, err1 := os.Stat(extractFilename)
				if os.IsNotExist(err1) {
					fmt.Println("[+] Downloading: "+fileUrl)
					err := DownloadFile(extractFilename, fileUrl)
					if err!=nil {
						if strings.Contains(err.Error(),"no such host") {
							fmt.Println("[*] Error: ",err)
							os.Exit(3)
						} else {
							fmt.Println("[*] Error: ",err)
						}						
					}
				}
				lines, err2 := readLines(extractFilename)
				if err2==nil {
					for _, v := range lines {
						v=strings.TrimSpace(v)
						if len(v)>0 {
							pathList = append(pathList, v)
						}
					}		
				} else {
					fmt.Println(err2)
				}

			}
		}		
		if Pathsource=="exploitdb-asp" {
			fileUrl := "https://raw.githubusercontent.com/milo2012/pathbrute/master/exploitdb_asp.txt"
			tokens := strings.Split(fileUrl,"/")
			var extractFilename=tokens[len(tokens)-1]
			if argv.Updatemode {
				statusCheck:=checkAndUpdate(fileUrl)
				if statusCheck==false {
					fmt.Println("[+] No update required for "+extractFilename)
				} else {
					fmt.Println("[+] Latest version of "+extractFilename+" has been downloaded")
				}
				lines, err2 := readLines(extractFilename)
				if err2==nil {
					for _, v := range lines {
						v=strings.TrimSpace(v)
						if len(v)>0 {
							pathList = append(pathList, v)
						}
					}		
				} else {
					fmt.Println(err2)
				}
			} else {
				_, err1 := os.Stat(extractFilename)
				if os.IsNotExist(err1) {
					fmt.Println("[+] Downloading: "+fileUrl)
					err := DownloadFile(extractFilename, fileUrl)
					if err!=nil {
						if strings.Contains(err.Error(),"no such host") {
							fmt.Println("[*] Error: ",err)
							os.Exit(3)
						} else {
							fmt.Println("[*] Error: ",err)
						}						
					}
				}
				lines, err2 := readLines(extractFilename)
				if err2==nil {
					for _, v := range lines {
						v=strings.TrimSpace(v)
						if len(v)>0 {
							pathList = append(pathList, v)
						}
					}		
				} else {
					fmt.Println(err2)
				}

			}
		}		
		if Pathsource=="exploitdb-aspx" {
			fileUrl := "https://raw.githubusercontent.com/milo2012/pathbrute/master/exploitdb_aspx.txt"
			tokens := strings.Split(fileUrl,"/")
			var extractFilename=tokens[len(tokens)-1]
			if argv.Updatemode {
				statusCheck:=checkAndUpdate(fileUrl)
				if statusCheck==false {
					fmt.Println("[+] No update required for "+extractFilename)
				} else {
					fmt.Println("[+] Latest version of "+extractFilename+" has been downloaded")
				}
				lines, err2 := readLines(extractFilename)
				if err2==nil {
					for _, v := range lines {
						v=strings.TrimSpace(v)
						if len(v)>0 {
							pathList = append(pathList, v)
						}
					}		
				} else {
					fmt.Println(err2)
				}
			} else {
				_, err1 := os.Stat(extractFilename)
				if os.IsNotExist(err1) {
					fmt.Println("[+] Downloading: "+fileUrl)
					err := DownloadFile(extractFilename, fileUrl)
					if err!=nil {
						if strings.Contains(err.Error(),"no such host") {
							fmt.Println("[*] Error: ",err)
							os.Exit(3)
						} else {
							fmt.Println("[*] Error: ",err)
						}						
					}
				}
				lines, err2 := readLines(extractFilename)
				if err2==nil {
					for _, v := range lines {
						v=strings.TrimSpace(v)
						if len(v)>0 {
							pathList = append(pathList, v)
						}
					}		
				} else {
					fmt.Println(err2)
				}

			}
		}		
		if Pathsource=="exploitdb-cfm" {
			fileUrl := "https://raw.githubusercontent.com/milo2012/pathbrute/master/exploitdb_cfm.txt"
			tokens := strings.Split(fileUrl,"/")
			var extractFilename=tokens[len(tokens)-1]
			if argv.Updatemode {
				statusCheck:=checkAndUpdate(fileUrl)
				if statusCheck==false {
					fmt.Println("[+] No update required for "+extractFilename)
				} else {
					fmt.Println("[+] Latest version of "+extractFilename+" has been downloaded")
				}
				lines, err2 := readLines(extractFilename)
				if err2==nil {
					for _, v := range lines {
						v=strings.TrimSpace(v)
						if len(v)>0 {
							pathList = append(pathList, v)
						}
					}		
				} else {
					fmt.Println(err2)
				}
			} else {
				_, err1 := os.Stat(extractFilename)
				if os.IsNotExist(err1) {
					fmt.Println("[+] Downloading: "+fileUrl)
					err := DownloadFile(extractFilename, fileUrl)
					if err!=nil {
						if strings.Contains(err.Error(),"no such host") {
							fmt.Println("[*] Error: ",err)
							os.Exit(3)
						} else {
							fmt.Println("[*] Error: ",err)
						}						
					}
				}
				lines, err2 := readLines(extractFilename)
				if err2==nil {
					for _, v := range lines {
						v=strings.TrimSpace(v)
						if len(v)>0 {
							pathList = append(pathList, v)
						}
					}		
				} else {
					fmt.Println(err2)
				}

			}
		}	
		if Pathsource=="exploitdb-cgi" {
			fileUrl := "https://raw.githubusercontent.com/milo2012/pathbrute/master/exploitdb_cgi.txt"
			tokens := strings.Split(fileUrl,"/")
			var extractFilename=tokens[len(tokens)-1]
			if argv.Updatemode {
				statusCheck:=checkAndUpdate(fileUrl)
				if statusCheck==false {
					fmt.Println("[+] No update required for "+extractFilename)
				} else {
					fmt.Println("[+] Latest version of "+extractFilename+" has been downloaded")
				}
				lines, err2 := readLines(extractFilename)
				if err2==nil {
					for _, v := range lines {
						v=strings.TrimSpace(v)
						if len(v)>0 {
							pathList = append(pathList, v)
						}
					}		
				} else {
					fmt.Println(err2)
				}
			} else {
				_, err1 := os.Stat(extractFilename)
				if os.IsNotExist(err1) {
					fmt.Println("[+] Downloading: "+fileUrl)
					err := DownloadFile(extractFilename, fileUrl)
					if err!=nil {
						if strings.Contains(err.Error(),"no such host") {
							fmt.Println("[*] Error: ",err)
							os.Exit(3)
						} else {
							fmt.Println("[*] Error: ",err)
						}						
					}
				}
				lines, err2 := readLines(extractFilename)
				if err2==nil {
					for _, v := range lines {
						v=strings.TrimSpace(v)
						if len(v)>0 {
							pathList = append(pathList, v)
						}
					}		
				} else {
					fmt.Println(err2)
				}

			}
		}	
		if Pathsource=="exploitdb-cfm" {
			fileUrl := "https://raw.githubusercontent.com/milo2012/pathbrute/master/exploitdb_cfm.txt"
			tokens := strings.Split(fileUrl,"/")
			var extractFilename=tokens[len(tokens)-1]
			if argv.Updatemode {
				statusCheck:=checkAndUpdate(fileUrl)
				if statusCheck==false {
					fmt.Println("[+] No update required for "+extractFilename)
				} else {
					fmt.Println("[+] Latest version of "+extractFilename+" has been downloaded")
				}
				lines, err2 := readLines(extractFilename)
				if err2==nil {
					for _, v := range lines {
						v=strings.TrimSpace(v)
						if len(v)>0 {
							pathList = append(pathList, v)
						}
					}		
				} else {
					fmt.Println(err2)
				}
			} else {
				_, err1 := os.Stat(extractFilename)
				if os.IsNotExist(err1) {
					fmt.Println("[+] Downloading: "+fileUrl)
					err := DownloadFile(extractFilename, fileUrl)
					if err!=nil {
						if strings.Contains(err.Error(),"no such host") {
							fmt.Println("[*] Error: ",err)
							os.Exit(3)
						} else {
							fmt.Println("[*] Error: ",err)
						}						
					}
				}
				lines, err2 := readLines(extractFilename)
				if err2==nil {
					for _, v := range lines {
						v=strings.TrimSpace(v)
						if len(v)>0 {
							pathList = append(pathList, v)
						}
					}		
				} else {
					fmt.Println(err2)
				}

			}
		}	
		if Pathsource=="exploitdb-jsp" {
			fileUrl := "https://raw.githubusercontent.com/milo2012/pathbrute/master/exploitdb_jsp.txt"
			tokens := strings.Split(fileUrl,"/")
			var extractFilename=tokens[len(tokens)-1]
			if argv.Updatemode {
				statusCheck:=checkAndUpdate(fileUrl)
				if statusCheck==false {
					fmt.Println("[+] No update required for "+extractFilename)
				} else {
					fmt.Println("[+] Latest version of "+extractFilename+" has been downloaded")
				}
				lines, err2 := readLines(extractFilename)
				if err2==nil {
					for _, v := range lines {
						v=strings.TrimSpace(v)
						if len(v)>0 {
							pathList = append(pathList, v)
						}
					}		
				} else {
					fmt.Println(err2)
				}
			} else {
				_, err1 := os.Stat(extractFilename)
				if os.IsNotExist(err1) {
					fmt.Println("[+] Downloading: "+fileUrl)
					err := DownloadFile(extractFilename, fileUrl)
					if err!=nil {
						if strings.Contains(err.Error(),"no such host") {
							fmt.Println("[*] Error: ",err)
							os.Exit(3)
						} else {
							fmt.Println("[*] Error: ",err)
						}						
					}
				}
				lines, err2 := readLines(extractFilename)
				if err2==nil {
					for _, v := range lines {
						v=strings.TrimSpace(v)
						if len(v)>0 {
							pathList = append(pathList, v)
						}
					}		
				} else {
					fmt.Println(err2)
				}

			}
		}	
		if Pathsource=="exploitdb-perl" {
			fileUrl := "https://raw.githubusercontent.com/milo2012/pathbrute/master/exploitdb_perl.txt"
			tokens := strings.Split(fileUrl,"/")
			var extractFilename=tokens[len(tokens)-1]
			if argv.Updatemode {
				statusCheck:=checkAndUpdate(fileUrl)
				if statusCheck==false {
					fmt.Println("[+] No update required for "+extractFilename)
				} else {
					fmt.Println("[+] Latest version of "+extractFilename+" has been downloaded")
				}
				lines, err2 := readLines(extractFilename)
				if err2==nil {
					for _, v := range lines {
						v=strings.TrimSpace(v)
						if len(v)>0 {
							pathList = append(pathList, v)
						}
					}		
				} else {
					fmt.Println(err2)
				}
			} else {
				_, err1 := os.Stat(extractFilename)
				if os.IsNotExist(err1) {
					fmt.Println("[+] Downloading: "+fileUrl)
					err := DownloadFile(extractFilename, fileUrl)
					if err!=nil {						
						if strings.Contains(err.Error(),"no such host") {
							fmt.Println("[*] Error: ",err)
							os.Exit(3)
						} else {
							fmt.Println("[*] Error: ",err)
						}						
					}
				}
				lines, err2 := readLines(extractFilename)
				if err2==nil {
					for _, v := range lines {
						v=strings.TrimSpace(v)
						if len(v)>0 {
							pathList = append(pathList, v)
						}
					}		
				} else {
					fmt.Println(err2)
				}

			}
		}	
		if Pathsource=="exploitdb-php" {
			fileUrl := "https://raw.githubusercontent.com/milo2012/pathbrute/master/exploitdb_php.txt"
			tokens := strings.Split(fileUrl,"/")
			var extractFilename=tokens[len(tokens)-1]
			if argv.Updatemode {
				statusCheck:=checkAndUpdate(fileUrl)
				if statusCheck==false {
					fmt.Println("[+] No update required for "+extractFilename)
				} else {
					fmt.Println("[+] Latest version of "+extractFilename+" has been downloaded")
				}
				lines, err2 := readLines(extractFilename)
				if err2==nil {
					for _, v := range lines {
						v=strings.TrimSpace(v)
						if len(v)>0 {
							pathList = append(pathList, v)
						}
					}		
				} else {
					fmt.Println(err2)
				}
			} else {
				_, err1 := os.Stat(extractFilename)
				if os.IsNotExist(err1) {
					fmt.Println("[+] Downloading: "+fileUrl)
					err := DownloadFile(extractFilename, fileUrl)
					if err!=nil {
						if strings.Contains(err.Error(),"no such host") {
							fmt.Println("[*] Error: ",err)
							os.Exit(3)
						} else {
							fmt.Println("[*] Error: ",err)
						}						
					}
				}
				lines, err2 := readLines(extractFilename)
				if err2==nil {
					for _, v := range lines {
						v=strings.TrimSpace(v)
						if len(v)>0 {
							pathList = append(pathList, v)
						}
					}		
				} else {
					fmt.Println(err2)
				}

			}
		}	
		if Pathsource=="exploitdb-others" {
			fileUrl := "https://raw.githubusercontent.com/milo2012/pathbrute/master/exploitdb-others.txt"
			tokens := strings.Split(fileUrl,"/")
			var extractFilename=tokens[len(tokens)-1]
			if argv.Updatemode {
				statusCheck:=checkAndUpdate(fileUrl)
				if statusCheck==false {
					fmt.Println("[+] No update required for "+extractFilename)
				} else {
					fmt.Println("[+] Latest version of "+extractFilename+" has been downloaded")
				}
				lines, err2 := readLines(extractFilename)
				if err2==nil {
					for _, v := range lines {
						v=strings.TrimSpace(v)
						if len(v)>0 {
							pathList = append(pathList, v)
						}
					}		
				} else {
					fmt.Println(err2)
				}
			} else {
				_, err1 := os.Stat(extractFilename)
				if os.IsNotExist(err1) {
					fmt.Println("[+] Downloading: "+fileUrl)
					err := DownloadFile(extractFilename, fileUrl)
					if err!=nil {
						if strings.Contains(err.Error(),"no such host") {
							fmt.Println("[*] Error: ",err)
							os.Exit(3)
						} else {
							fmt.Println("[*] Error: ",err)
						}						
					}
				}
				lines, err2 := readLines(extractFilename)
				if err2==nil {
					for _, v := range lines {
						v=strings.TrimSpace(v)
						if len(v)>0 {
							pathList = append(pathList, v)
						}
					}		
				} else {
					fmt.Println(err2)
				}

			}
		}	
		if Pathsource=="SecLists" {
			fileUrl := "https://raw.githubusercontent.com/danielmiessler/SecLists/master/Discovery/Web-Content/common.txt"
			tokens := strings.Split(fileUrl,"/")
			var extractFilename=tokens[len(tokens)-1]
			if argv.Updatemode {
				statusCheck:=checkAndUpdate(fileUrl)
				if statusCheck==false {
					fmt.Println("[+] No update required for "+extractFilename)
				} else {
					fmt.Println("[+] Latest version of "+extractFilename+" has been downloaded")
				}
				lines, err2 := readLines(extractFilename)
				if err2==nil {
					for _, v := range lines {
						v=strings.TrimSpace(v)
						if len(v)>0 {
							pathList = append(pathList, v)
						}
					}		
				} else {
					fmt.Println(err2)
				}
			} else {
				_, err1 := os.Stat(extractFilename)
				if os.IsNotExist(err1) {
					fmt.Println("[+] Downloading: "+fileUrl)
					err := DownloadFile(extractFilename, fileUrl)
					if err!=nil {
						if strings.Contains(err.Error(),"no such host") {
							fmt.Println("[*] Error: ",err)
							os.Exit(3)
						} else {
							fmt.Println("[*] Error: ",err)
						}						
					}
				}
				lines, err2 := readLines(extractFilename)
				if err2==nil {
					for _, v := range lines {
						v=strings.TrimSpace(v)
						if len(v)>0 {
							pathList = append(pathList, v)
						}
					}		
				} else {
					fmt.Println(err2)
				}

			}
		}
		if Pathsource=="RobotsDisallowed" {
			fileUrl := "https://raw.githubusercontent.com/danielmiessler/RobotsDisallowed/master/Top100000-RobotsDisallowed.txt"
			tokens := strings.Split(fileUrl,"/")
			var extractFilename=tokens[len(tokens)-1]
			if argv.Updatemode {
				statusCheck:=checkAndUpdate(fileUrl)
				if statusCheck==false {
					fmt.Println("[+] No update required for "+extractFilename)
				} else {
					fmt.Println("[+] Latest version of "+extractFilename+" has been downloaded")
				}
				lines, err2 := readLines(extractFilename)
				if err2==nil {
					for _, v := range lines {
						v=strings.TrimSpace(v)
						if len(v)>0 {
							pathList = append(pathList, v)
						}
					}		
				} else {
					fmt.Println(err2)
				}
			} else {
				_, err1 := os.Stat(extractFilename)
				if os.IsNotExist(err1) {
					fmt.Println("[+] Downloading: "+fileUrl)
					err := DownloadFile(extractFilename, fileUrl)
					if err!=nil {
						if strings.Contains(err.Error(),"no such host") {
							fmt.Println("[*] Error: ",err)
							os.Exit(3)
						} else {
							fmt.Println("[*] Error: ",err)
						}						
					}
				}
				lines, err2 := readLines(extractFilename)
				if err2==nil {
					for _, v := range lines {
						v=strings.TrimSpace(v)
						if len(v)>0 {
							pathList = append(pathList, v)
						}
					}		
				} else {
					fmt.Println(err2)
				}
			}
		}
		if len(argv.URLpath)<1 && len(argv.Filename)<1 {
			fmt.Println("[!] Please set the -U or the -u argument")
			os.Exit(3)
		} else {
			if len(argv.Filename)>0 {
				_, err := os.Stat(filename1)
				if os.IsNotExist(err) {
					fmt.Printf("[*] File %s not exists\n", filename1)
					os.Exit(3)
				}
				lines, err := readLines(filename1)
				if err!=nil {
					fmt.Println("Error: ",err)
				} else {
					for _, v := range lines {
						if strings.Contains(v,"http") {
							contentList = append(contentList, v)
						} else {
							if len(v)>0 {
								contentList = append(contentList, "https://"+v)
								contentList = append(contentList, "http://"+v)
							}
						}
						//fmt.Println("https://"+v)
					}	
				}
			} else {
				if strings.Contains(argv.URLpath,"http") {
					contentList = append(contentList, argv.URLpath)
				} else {
					if len(argv.URLpath)>0 {
						contentList = append(contentList, "https://"+argv.URLpath)
						contentList = append(contentList, "http://"+argv.URLpath)
					}
				}
			}
		}

		var contentList1 []string
  	    for _, v := range contentList {
			if strings.HasSuffix(v,":443") {
				v=v[0:len(v)-4]
				v=strings.TrimSpace(v)
				if len(v)>0 {
					if !stringInSlice(v,contentList1) {
						contentList1 = append(contentList1, v)
					}
				}
			} else {
				v=strings.TrimSpace(v)
				if len(v)>0 {
					contentList1 = append(contentList1, v)
				}
			}			
  	    }
		contentList=contentList1
		//_ = contentList1

		if argv.CMSmode {
			CMSmode = true
			pathList = append(pathList, "/wp-links-opml.php")
		    for _, v := range wpFileList {
		    	pathList = append(pathList,v)
		    }			
		    for _, v := range joomlaFileList {
		    	pathList = append(pathList,v)
		    }
		    for _, v := range drupalFileList {
		    	pathList = append(pathList,v)
		    }
		} 
		
		var finalList []string


		//sigs1 := make(chan os.Signal, 1)
		//signal.Notify(sigs1, syscall.SIGINT, syscall.SIGTERM)

		urlChan := make(chan string)
		if intelligentMode==true {
			var wg1 sync.WaitGroup
			wg1.Add(workersCount)
	
			for i := 0; i < workersCount; i++ {
				go func() {
					//sig := <-sigs1
					testFakePath(urlChan)
					wg1.Done()
					//fmt.Println(sig)
					//done <- true
				}()
			}

			fmt.Println("[*] Pre-Bruteforce Checks")
			log.Println("[*] Pre-Bruteforce Checks")
			completed := 0
			for _, each := range contentList {
				urlChan <- each+"/NonExistence"
				completed++
			}
			close(urlChan)    
			for {
				time.Sleep(10 * time.Millisecond)
				if len(contentList)==int(currentFakeCount) {
					break
				}
			}
		}

		var contentList2 []string
		for _, x := range contentList {
			tmpFound:=false
			if argv.Skipmode {
				for _, v := range blacklistList {
					if strings.Contains(x,v) {
						tmpFound=true
						fmt.Println("[Skip] "+x)
					}
				}	
			}
			if tmpFound==false {
				contentList2 = append(contentList2,x)
			}	
		}		

		if SpreadMode==false {
			for _, x := range contentList2 {
			  for _, v := range pathList {
				url := x      		
				path := v
				newUrl := ""
				if strings.HasSuffix(url,"/") {
					url=url[0:len(url)-1]
				}			
				if strings.HasPrefix(path,"/") {
					newUrl = url+path
				} else {		
					newUrl = url+"/"+path
				}
				finalList = append(finalList, newUrl)
			  }
			}
		} else {
 	 	    for _, v := range pathList {
			  for _, x := range contentList2 {
				url := x      		
				path := v
				newUrl := ""
				if strings.HasSuffix(url,"/") {
					url=url[0:len(url)-1]
				}			
				if strings.HasPrefix(path,"/") {
					newUrl = url+path
				} else {		
					newUrl = url+"/"+path
				}
				finalList = append(finalList, newUrl)
			  }
			}
		}

		/*var wg sync.WaitGroup
		urlChan = make(chan string)
		wg.Add(workersCount)
	
		for i := 0; i < workersCount; i++ {
			go func() {
				getUrlWorker(urlChan)
				wg.Done()
			}()
		}*/

		totalListCount=len(finalList)

		fmt.Println("\n[*] Testing URI Paths: (Total: "+strconv.Itoa(totalListCount)+")")		
		log.Printf("\n[*] Testing URI Paths")
		if totalListCount==0 {
			fmt.Println("[-] There are no URI paths to be tested.")
			os.Exit(3)
		} 
		if ContinueNum>totalListCount {
			fmt.Println("For the -r option, you must enter a value smaller than "+strconv.Itoa(totalListCount))
			os.Exit(3)
		} 
		//real uripaths
		completed1 := 0
		
		async := nasync.New(workersCount,workersCount)
		defer async.Close()

		for _, each := range finalList {
			if ContinueNum==0 || ContinueNum<completed1+1 {	
				async.Do(testURL,each+" | "+strconv.Itoa(completed1+1))
			}
			completed1++
		}
		//close(urlChan) 
		
		for {			
			time.Sleep(1 * time.Millisecond)
			if reachedTheEnd==true && completedCount==len(finalList){
				break
			} 
		}
			
		if CMSmode==true {
			for _, v := range tmpResultList {
				var wpVer = ""
				timeout := time.Duration(time.Duration(timeoutSec) * time.Second)
				client := http.Client{
					Timeout: timeout,
					CheckRedirect: redirectPolicyFunc,		
				}

				if proxyMode==true {
					url_i := url.URL{}
					url_proxy, _ := url_i.Parse(proxy_addr)
					http.DefaultTransport.(*http.Transport).Proxy = http.ProxyURL(url_proxy)
				}
				if strings.HasSuffix(v[0],"/administrator/language/en-GB/en-GB.xml") || strings.HasSuffix(v[0],"/administrator/manifests/files/joomla.xml") {
					http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

					req, err := http.NewRequest("GET", v[0], nil)
					req.Header.Add("User-Agent", userAgent)
					req.SetBasicAuth(identUser, identPass)
					resp, err := client.Do(req)		
					if resp!=nil{					
						defer resp.Body.Close()
					}
					//resp, err := client.Get(v[0])
					if err==nil {
						body, err := ioutil.ReadAll(resp.Body)
						if err==nil {
							bodyStr := BytesToString(body)
							if strings.Contains(bodyStr,"_Incapsula_Resource") {
								wpVer="- Protected by Incapsula"
							} else {
								s := strings.Split(bodyStr,"\n")
								for _, v1 := range s {

									if strings.Contains(v1,"<version>") {
										v1=strings.Replace(v1,"</version>","",1)
										v1=strings.Replace(v1,"<version>","",1)
										v1=strings.TrimSpace(v1)
										wpVer = v1
									}
								}
							}
						}
						v[0]=strings.Replace(v[0],"/administrator/language/en-GB/en-GB.xml","",1)
						v[0]=strings.Replace(v[0],"/administrator/manifests/files/joomla.xml","",1)					
						if len(wpVer)>0 {
							var a = color.BlueString("\n[Found] ")+v[0]+color.BlueString(" [Joomla "+wpVer+"]")
							tmpResultList1 = append(tmpResultList1, a)
						}
					}
				}
				if strings.Contains(v[0],"/CHANGELOG.txt") {
					if proxyMode==true {
						url_i := url.URL{}
						url_proxy, _ := url_i.Parse(proxy_addr)
						http.DefaultTransport.(*http.Transport).Proxy = http.ProxyURL(url_proxy)
					}
					http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
					req, _ := http.NewRequest("GET", v[0], nil)
					req.Header.Add("User-Agent", userAgent)
					req.SetBasicAuth(identUser, identPass)
					resp, err := client.Do(req)		
					if resp!=nil{					
						defer resp.Body.Close()
					}
					//resp, err := client.Get(v[0])
					if err==nil {
						body, err := ioutil.ReadAll(resp.Body)
						if err==nil {
							bodyStr := BytesToString(body)
							s := strings.Split(bodyStr,"\n")
							var tmpFound = false
							for _, v1 := range s {
								if tmpFound==false {
									if strings.Contains(v1,"Drupal ") {
										v1=strings.TrimSpace(v1)
										wpVer = strings.Split(v1,",")[0]
										tmpFound=true
									}
								}
							}
						}
						v[0]=strings.Replace(v[0],"/CHANGELOG.txt","",1)					
						if len(wpVer)>0 {
							var a = color.BlueString("\n[Found] ")+v[0]+color.BlueString(" ["+wpVer+"]")
							tmpResultList1 = append(tmpResultList1, a)
						}
					}
				}				

				if strings.Contains(v[0],"/readme.html") {
					if proxyMode==true {
						url_i := url.URL{}
						url_proxy, _ := url_i.Parse(proxy_addr)
						http.DefaultTransport.(*http.Transport).Proxy = http.ProxyURL(url_proxy)
					}
					http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

					req, _ := http.NewRequest("GET", v[0], nil)
					req.Header.Add("User-Agent", userAgent)
					req.SetBasicAuth(identUser, identPass)
					resp, err := client.Do(req)		
					if resp!=nil{					
						defer resp.Body.Close()
					}
					//resp, err := client.Get(v[0])
					if err==nil {
						body, err := ioutil.ReadAll(resp.Body)
						if err==nil {
							bodyStr := BytesToString(body)
							s := strings.Split(bodyStr,"\n")
							for _, v1 := range s {
								if strings.Contains(v1,"<br /> Versão ") {
									v1=removeCharacters(v1,"<br /> Versão ")
									v1=strings.TrimSpace(v1)
									wpVer = v1
								}
								if strings.Contains(v1,"<br /> Version ") {
									v1=removeCharacters(v1,"<br /> Version ")
									v1=strings.TrimSpace(v1)
									wpVer = v1
								}
							}
						}
					}
					v[0]=strings.Replace(v[0],"/readme.html","",1)
					if len(wpVer)>0 {
						var a = color.BlueString("\n[Found] ")+v[0]+color.BlueString(" [Wordpress "+wpVer+"]")
						tmpResultList1 = append(tmpResultList1, a)
					}		
				}
				//if strings.HasPrefix(v[3],"Links for ") {			
				if strings.Contains(v[0],"/wp-links-opml.php") {
					if proxyMode==true {
						url_i := url.URL{}
						url_proxy, _ := url_i.Parse(proxy_addr)
						http.DefaultTransport.(*http.Transport).Proxy = http.ProxyURL(url_proxy)
					}
					http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

					req, _ := http.NewRequest("GET", v[0], nil)
					req.Header.Add("User-Agent", userAgent)
					req.SetBasicAuth(identUser, identPass)
					resp, err := client.Do(req)		
					if resp!=nil{					
						defer resp.Body.Close()
					}
					if err==nil {
						body, err := ioutil.ReadAll(resp.Body)
						if err==nil {
							bodyStr := BytesToString(body)
							s := strings.Split(bodyStr,"\n")
							var tmpFound=false
							for _, v1 := range s {
								if strings.Contains(v1,"Links for ") {
									tmpFound=true
								}
								if strings.Contains(v1," generator=\"") {
									v1=removeCharacters(v1,"<!--  generator=\"WordPress\"/")
									v1=removeCharacters(v1,"<!-- generator=\"WordPress\"/")
									v1=removeCharacters(v1,"\" -->")
									v1=strings.TrimSpace(v1)
									wpVer = v1
								} 
							}
							if tmpFound==true && len(wpVer)<1 {
								wpVer="(Unknown)"								
							}
						}
					}
					
					v[0]=strings.Replace(v[0],"/wp-links-opml.php","",1)
					if len(wpVer)>0 {
						var a = color.BlueString("\n[Found] ")+v[0]+color.BlueString(" [Wordpress "+wpVer+"]")
						tmpResultList1 = append(tmpResultList1, a)
					}		
				}
			}
		} else {
			for _, v := range tmpResultList {
				if !stringInSlice(v[0],tmpResultList1) {
					tmpResultList1 = append(tmpResultList1, v[0])
				}
			}

			var tmpResultList2 []string	

			sort.Strings(tmpResultList1)
			for _, v := range tmpResultList1 {
				u, err := url.Parse(v)
				if err==nil {
					if len(u.Path)>0 {
						tmpResultList2 = append(tmpResultList2,v)
					}
				}
			}					
			
			if len(tmpResultList2)<1 {
				fmt.Printf("\n[*] No results found")
				log.Printf("\n[*] No results found")
			} else {
				fmt.Printf("\n")
				log.Printf("\n")
				async := nasync.New(workersCount,workersCount)
				defer async.Close()
		
				var wg sync.WaitGroup
				urlChan = make(chan string)
				wg.Add(workersCount)
				for i := 0; i < workersCount; i++ {
					go func() {	
						checkURL(urlChan)
						wg.Done()
					}()
				}		
				for _, each := range tmpResultList2 {
					async.Do(checkURL1,each)
					urlChan <- each
				}
				close(urlChan)  
				wg.Wait()				
			}
			for {	
				if reachedTheEnd1==true {
					break
				}
				if int(currentListCount1)>=len(tmpResultList2) {
					reachedTheEnd1=true
				} 
			}
		}
		

		if CMSmode==true {
			var joomlaKBList [][]string	
			var wpKBList [][]string	
			var drupalKBList [][]string	

			var a = [][]string{{"joomla","3.7.0","Joomla Component Fields SQLi Remote Code Execution","exploit/unix/webapp/joomla_comfields_sqli_rce"}}
			joomlaKBList = append(joomlaKBList,a...)
			var b = [][]string{{"joomla","3.4.4-3.6.3","Joomla Account Creation and Privilege Escalation","auxiliary/admin/http/joomla_registration_privesc"}}
			joomlaKBList = append(joomlaKBList,b...)
			var c = [][]string{{"joomla","1.5.0-3.4.5","Joomla HTTP Header Unauthenticated Remote Code Execution","exploit/multi/http/joomla_http_header_rce"}}
			joomlaKBList = append(joomlaKBList,c...)
			var d = [][]string{{"joomla","3.2-3.4.4","Joomla Content History SQLi Remote Code Execution","exploit/unix/webapp/joomla_contenthistory_sqli_rce"}}
			joomlaKBList = append(joomlaKBList,d...)
			var e = [][]string{{"joomla","2.5.0-2.5.13,3.0.0-3.1.4","Joomla Media Manager File Upload Vulnerability","exploit/unix/webapp/joomla_media_upload_exec"}}
			joomlaKBList = append(joomlaKBList,e...)

			a = [][]string{{"wordpress","4.6","WordPress PHPMailer Host Header Command Injection","exploit/unix/webapp/wp_phpmailer_host_header"}}
			wpKBList = append(wpKBList,a...)
			a = [][]string{{"wordpress","4.5.1","WordPress Same-Origin Method Execution (SOME)","https://gist.github.com/cure53/09a81530a44f6b8173f545accc9ed07e (http://example.com/wp-includes/js/plupload/plupload.flash.swf?target%g=alert&uid%g=hello&)"}}
			wpKBList = append(wpKBList,a...)
			b = [][]string{{"wordpress","4.7-4.7.1","WordPress REST API Content Injection","auxiliary/dos/http/wordpress_long_password_dos"}}
			wpKBList = append(wpKBList,b...)
			c = [][]string{{"wordpress","3.7.5,3.9-3.9.3,4.0-4.0.1","WordPress Long Password DoS",""}}
			wpKBList = append(wpKBList,c...)
			d = [][]string{{"wordpress","3.5-3.9.2","Wordpress XMLRPC DoS","auxiliary/dos/http/wordpress_xmlrpc_dos"}}
			wpKBList = append(wpKBList,d...)
			e = [][]string{{"wordpress","0-1.5.1.3","WordPress cache_lastpostdate Arbitrary Code Execution","exploit/unix/webapp/wp_lastpost_exec"}}
			wpKBList = append(wpKBList,e...)
			var f = [][]string{{"wordpress","0-4.4.1","Wordpress XML-RPC system.multicall Credential Collector","auxiliary/scanner/http/wordpress_multicall_creds"}}
			wpKBList = append(wpKBList,f...)
			var g = [][]string{{"wordpress","0-4.6","WordPress Traversal Directory DoS","auxiliary/dos/http/wordpress_directory_traversal_dos"}}
			wpKBList = append(wpKBList,g...)
			
			a = [][]string{{"drupal","7.0,7.31","Drupal HTTP Parameter Key/Value SQL Injection","exploit/multi/http/drupal_drupageddon"}}
			drupalKBList = append(drupalKBList,a...)
			b = [][]string{{"drupal","7.15,7.2","PHP XML-RPC Arbitrary Code Execution","exploit/unix/webapp/php_xmlrpc_eval"}}
			drupalKBList = append(drupalKBList,b...)
			c = [][]string{{"drupal","7.0-7.56,8.0<8.3.9,8.4.0<8.4.6,8.5.0-8.5.1","CVE-2018-7600 / SA-CORE-2018-002","https://github.com/a2u/CVE-2018-7600"}}
			drupalKBList = append(drupalKBList,c...)
			
			RemoveDuplicates(&tmpResultList1)
			sort.Strings(tmpResultList1)
			if len(tmpResultList1)>0 {
				fmt.Printf("\n")
			}
			for _, v1 := range tmpResultList1 {
				fmt.Printf("%s\n",v1)
				if strings.Contains(v1,"Joomla") {
					tmpSplit1 :=strings.Split(v1,"[Joomla ")
					tmpSplit2 :=strings.Split(tmpSplit1[1],"]")
					selectedVer := tmpSplit2[0]	
					for _, v := range joomlaKBList {
						if strings.Contains(v[1],",") {
							s := strings.Split(string(v[1]),",")
							for _, s1 := range s {
								if strings.Contains(s1,"-") {
									s2 := strings.Split(s1,"-")
									va0, _ := version.NewVersion(selectedVer)
									va1, _ := version.NewVersion(s2[0])
									va2, _ := version.NewVersion(s2[1])
									if va0.LessThan(va2) && va0.GreaterThan(va1) { 
										fmt.Printf("\n[Vuln] %s [%s]\n",v[2],v[3])
										log.Printf("[Vuln] 	%s [%s]\n",v[2],v[3])
									}
								} else if strings.Contains(s1,"<") {
									s2 := strings.Split(s1,"<")
									va0, _ := version.NewVersion(selectedVer)
									va1, _ := version.NewVersion(s2[0])
									va2, _ := version.NewVersion(s2[1])
									if va0.LessThan(va2) && va0.GreaterThan(va1) { 
										fmt.Printf("\n[Vuln] %s [%s]\n",v[2],v[3])
										log.Printf("[Vuln] %s [%s]\n",v[2],v[3])
									}
								} else { 
									va0, _ := version.NewVersion(selectedVer)
									va1, _ := version.NewVersion(s1)
									if va0.Equal(va1) {
										fmt.Printf("\n[Vuln] 	%s [%s]\n",v[2],v[3])
										log.Printf("[Vuln] 	%s [%s]\n",v[2],v[3])
									}
								}
							}	
						} else {
							if strings.Contains(v[1],"-") {
								s2 := strings.Split(v[1],"-")
								va0, _ := version.NewVersion(selectedVer)
								va1, _ := version.NewVersion(s2[0])
								va2, _ := version.NewVersion(s2[1])
								if va0.LessThan(va2) && va0.GreaterThan(va1) { 
									fmt.Printf("\n[Vuln] 	%s [%s]\n",v[2],v[3])
									log.Printf("[Vuln] 	%s [%s]\n",v[2],v[3])
								}
							} else if strings.Contains(v[1],"<") {
								s2 := strings.Split(v[1],"<")
								va0, _ := version.NewVersion(selectedVer)
								va1, _ := version.NewVersion(s2[0])
								va2, _ := version.NewVersion(s2[1])
								if va0.LessThan(va2) && va0.GreaterThan(va1) { 
									fmt.Printf("\n[Vuln] 	%s [%s]\n",v[2],v[3])
									log.Printf("[Vuln] 	%s [%s]\n",v[2],v[3])
								}
							} else { 
								va0, _ := version.NewVersion(selectedVer)
								va1, _ := version.NewVersion(v[1])
								if va0.Equal(va1) {
									fmt.Printf("\n[Vuln] 	%s [%s]\n",v[2],v[3])
									log.Printf("[Vuln] 	%s [%s]\n",v[2],v[3])
								}
							}

						}			
					}
				}					
				if strings.Contains(v1,"Wordpress") {
					tmpSplit1 :=strings.Split(v1,"[Wordpress ")
					tmpSplit2 :=strings.Split(tmpSplit1[1],"]")
					selectedVer := tmpSplit2[0]	
					if !strings.Contains(selectedVer,"(Unknown)") {
						for _, v := range wpKBList {
							if strings.Contains(v[1],",") {
								s := strings.Split(string(v[1]),",")
								for _, s1 := range s {
									if strings.Contains(s1,"-") {
										s2 := strings.Split(s1,"-")
										va0, _ := version.NewVersion(selectedVer)
										va1, _ := version.NewVersion(s2[0])
										va2, _ := version.NewVersion(s2[1])
										if va0.LessThan(va2) && va0.GreaterThan(va1) { 
											fmt.Printf("\n[Vuln] 	%s [%s]\n",v[2],v[3])
											log.Printf("[Vuln] 	%s [%s]\n",v[2],v[3])
										}
									} else if strings.Contains(s1,"<") {
										s2 := strings.Split(s1,"<")
										va0, _ := version.NewVersion(selectedVer)
										va1, _ := version.NewVersion(s2[0])
										va2, _ := version.NewVersion(s2[1])
										if va0.LessThan(va2) && va0.GreaterThan(va1) { 
											fmt.Printf("\n[Vuln] 	%s [%s]\n",v[2],v[3])
											log.Printf("[Vuln] 	%s [%s]\n",v[2],v[3])
										}
									} else {
										va0, _ := version.NewVersion(selectedVer)
										va1, _ := version.NewVersion(s1)
										if va0.Equal(va1) {
											fmt.Printf("\n[Vuln] 	%s [%s]\n",v[2],v[3])
											log.Printf("[Vuln] 	%s [%s]\n",v[2],v[3])
										}
									}
								}	
							} else {
								if strings.Contains(v[1],"-") {
									s2 := strings.Split(v[1],"-")
									va0, _ := version.NewVersion(selectedVer)
									va1, _ := version.NewVersion(s2[0])
									va2, _ := version.NewVersion(s2[1])
									if va0.LessThan(va2) && va0.GreaterThan(va1) { 
										fmt.Printf("\n[Vuln] 	%s [%s]\n",v[2],v[3])
										log.Printf("[Vuln] 	%s [%s]\n",v[2],v[3])
									}
								} else if strings.Contains(v[1],"<") {
									s2 := strings.Split(v[1],"<")
									va0, _ := version.NewVersion(selectedVer)
									va1, _ := version.NewVersion(s2[0])
									va2, _ := version.NewVersion(s2[1])
									if va0.LessThan(va2) && va0.GreaterThan(va1) { 
										fmt.Printf("\n[Vuln] 	%s [%s]\n",v[2],v[3])
										log.Printf("[Vuln] 	%s [%s]\n",v[2],v[3])
									}
								} else {
									va0, _ := version.NewVersion(selectedVer)
									va1, _ := version.NewVersion(v[1])
									if va0.Equal(va1) {
										fmt.Printf("\n[Vuln] 	%s [%s]\n",v[2],v[3])
										log.Printf("[Vuln] 	%s [%s]\n",v[2],v[3])
									}
								}

							}			
						}
					}
				}		
				if strings.Contains(v1,"Drupal") {
					tmpSplit1 :=strings.Split(v1,"[Drupal ")
					tmpSplit2 :=strings.Split(tmpSplit1[1],"]")
					selectedVer := tmpSplit2[0]	
					for _, v := range drupalKBList {
						if strings.Contains(v[1],",") {
							s := strings.Split(string(v[1]),",")
							for _, s1 := range s {
								if strings.Contains(s1,"-") {
									s2 := strings.Split(s1,"-")
									va0, _ := version.NewVersion(selectedVer)
									va1, _ := version.NewVersion(s2[0])
									va2, _ := version.NewVersion(s2[1])
									if va0.LessThan(va2) && va0.GreaterThan(va1) { 
										fmt.Printf("\n[Vuln] 	%s [%s]\n",v[2],v[3])
										log.Printf("[Vuln] 	%s [%s]\n",v[2],v[3])
									}
								} else if strings.Contains(s1,"<") {
									s2 := strings.Split(s1,"<")
									va0, _ := version.NewVersion(selectedVer)
									va1, _ := version.NewVersion(s2[0])
									va2, _ := version.NewVersion(s2[1])
									if va0.LessThan(va2) && va0.GreaterThan(va1) { 
										fmt.Printf("\n[Vuln] 	%s [%s]\n",v[2],v[3])
										log.Printf("[Vuln] 	%s [%s]\n",v[2],v[3])
									}
								} else {
									va0, _ := version.NewVersion(selectedVer)
									va1, _ := version.NewVersion(s1)
									if va0.Equal(va1) {
										fmt.Printf("\n[Vuln] 	%s [%s]\n",v[2],v[3])
										log.Printf("[Vuln] 	%s [%s]\n",v[2],v[3])
									}
								}
							}	
						} else {
							if strings.Contains(v[1],"-") {
								s2 := strings.Split(v[1],"-")
								va0, _ := version.NewVersion(selectedVer)
								va1, _ := version.NewVersion(s2[0])
								va2, _ := version.NewVersion(s2[1])
								if va0.LessThan(va2) && va0.GreaterThan(va1) { 
									fmt.Printf("\n[Vuln] 	%s [%s]\n",v[2],v[3])
									log.Printf("[Vuln] 	%s [%s]\n",v[2],v[3])
								}
							} else {
								va0, _ := version.NewVersion(selectedVer)
								va1, err := version.NewVersion(v[1])
								if err==nil {
									if va0.Equal(va1) {
										fmt.Printf("\n[Vuln] 	%s [%s]\n",v[2],v[3])
										log.Printf("[Vuln] 	%s [%s]\n",v[2],v[3])
									}
								}
							}

						}			
					}
				}				
			}
		}		
		return nil
	})
}
