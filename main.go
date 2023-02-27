package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	BingApi = "https://cn.bing.com/HPImageArchive.aspx?format=js&idx=0&n=10&nc=1612409408851&pid=hp&FORM=BEHPTB&uhd=1&uhdwidth=3840&uhdheight=2160"
	BingUrl = "https://cn.bing.com"
)

func main() {

	var everyDeskTop bool
	flag.BoolVar(&everyDeskTop, "every", true, "set desktop picture every day")
	today := time.Now().Format("2006-01-02")

	url, err := getBingImage()
	if err != nil {
		_ = fmt.Errorf("Error getting Bing image: %s\n", err)
	}

	filePath := filepath.Join(os.Getenv("HOME"), "Pictures", fmt.Sprintf("%s.jpg", today))

	if err := downloadFile(url, filePath); err != nil {
		fmt.Printf("Error downloading file: %s\n", err)
		return
	}

	if everyDeskTop {
		setDesktopPictureEvery(filePath)
	} else {
		setDesktopPicture(filePath)
	}

	fmt.Println("Desktop picture updated")
}

func getBingImage() (string, error) {
	req, err := http.NewRequest("GET", BingApi, nil)
	http.Header.Set(req.Header, "User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.114 Safari/537.36")

	if err != nil {
		return "", err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	// 读取响应体中的数据
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var response Response
	err = json.Unmarshal(data, &response)
	return BingUrl + response.Images[0].Url, nil
}

// downloadFile 下载文件
func downloadFile(url string, filePath string) error {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 将响应体写入文件
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func setDesktopPictureEvery(path string) {
	cmd := exec.Command("osascript", "-")
	cmd.Stdin = strings.NewReader(
		fmt.Sprintf(`
		tell application "System Events"
			set desktopCount to count of desktops
			repeat with i from 1 to desktopCount
				tell desktop i
					set picture to "%s"
				end tell
			end repeat
		end tell
	`, path))
	out, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(out))
}

func setDesktopPicture(path string) {
	cmd := exec.Command("osascript", "-e", fmt.Sprintf(`tell application "System Events" to set desktop picture of every desktop to "%s"`, path))
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error setting desktop picture: %s\n", err)
		return
	}
}
