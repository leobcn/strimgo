package main

import (
        "bufio"
        "fmt"
        "io/ioutil"
        "log"
        "strconv"

        "crypto/tls"
        "encoding/json"
        "net/http"
        "strings"
        "unicode/utf8"

        "os"
        "os/exec"
        "runtime"
        "sync"

        tbox "github.com/nsf/termbox-go"
)

var help string = "strimgo " + vid + "\n"+
         "Usage: strimgo [file]\n\n"+

         "R             refresh stream status\n"+
         "Up/Down k/j   select stream\n"+
         "Left/Right\n"+
         "h/l           scroll left/right\n"+
         "Home/End      go to start/end\n"+
         "Enter         run stream using medium,source quality\n"+
         "SHLMWA        use source/high/low/mobile/worst/audio\n"+
         "B             open stream page in browser\n"+
         "C             open chat popout in browser\n"+
         "V             open stream player popout in browser\n"+
         "Q/Escape      quit\n\n"+

         "Mouse: scroll up/down, left click (run stream using keyboard)\n"+
         "$HOME/.strimgo is used as default file, if unspecified\n"+
         "File consists of a list of channel names, separated by newlines\n"+
         "Loading random files is a terrible idea\n"+
         "Make sure there are no empty lines or trailing spaces\n"+
         "Livestreamer needs to be visible in PATH\n"+
         "https://github.com/muffindrake/strimgo"

var (
        path            string

        strm            []string
        stat            []bool
        index           []int
        map_s_g_t       map[string]*[2]string

        c_fmt           int
        g_fmt           int
        t_fmt           int
        dif             int

        w               int
        h               int
        cur             int
        scr_x           int
        scr_y           int
)

const (
        DEF     = tbox.ColorDefault
        REV     = tbox.ColorDefault | tbox.AttrReverse

        PAGE_DEFAULT            = iota
        PAGE_CHAT_POPOUT
        PAGE_VIDEO_POPOUT

        vid     = "v0"
        cid     = "strimgo_" + vid
        tapi    = "https://api.twitch.tv/kraken/streams/"
)

func main() {
        if runtime.GOOS == "windows" {
		path = os.ExpandEnv("%userprofile%/.strimgo")
        } else {path = os.ExpandEnv("$HOME/.strimgo")}

        switch {
        case len(os.Args) > 2:
                fmt.Println("Too many arguments.\n" + help)
                os.Exit(1)
        case len(os.Args) == 2:
                if path = os.Args[1]; path == "--help" {
                        fmt.Println(help)
                        os.Exit(0)
                }
        }

        file_done, env_done := make(chan bool), make(chan bool)

        go parse_strm(path, file_done)
        go parse_env(env_done)

        tr := &http.Transport{TLSClientConfig: &tls.Config{}}
        client := &http.Client{Transport: tr}

        if err := tbox.Init(); err != nil {panic(err)}
        defer tbox.Close()

        tbox.HideCursor()
        tbox.SetInputMode(tbox.InputEsc | tbox.InputMouse)
        w, h = tbox.Size()
        <-file_done
        if <-env_done {chk_stat(client)}

        draw_all()
main_loop:
        for{
        event_loop:
                switch e := tbox.PollEvent(); e.Type {
                case tbox.EventMouse:
                        switch e.Key {
                        case tbox.MouseWheelUp:
                                if cur == 0 {
                                        cur = len(index) - 1
                                        if len(index) > h {
                                                scr_y = len(index) - h + 1
                                        }
                                        break event_loop
                                } else {cur--}

                                if len(index) > h && scr_y != 0 {
                                        scr_y--
                                }

                                break event_loop
                        case tbox.MouseWheelDown:
                                if cur == len(index)-1 {
                                        cur = 0
                                        if len(index) > h {
                                                scr_y = 0
                                        }
                                        break event_loop
                                } else {cur++}

                                if len(index) > h && scr_y != len(index)-h+1 {
                                        scr_y++
                                }

                                break event_loop
                        case tbox.MouseLeft:
                                if e.MouseY < len(index) {
                                        if len(index) > h {
                                                cur = scr_y + e.MouseY
						if cur > len(index)-1 {
							cur = len(index) - 1
						}
						scr_y = cur - 1
						if scr_y < 0 {
							scr_y = 0
						}
					} else {
                                                cur = e.MouseY
                                        }
                                }
                        }
                case tbox.EventKey:
                        switch e.Key {
                        case tbox.KeyEsc:
                                break main_loop
                        case tbox.KeyEnter:
                                exc("medium,source")
                                break event_loop
                        case tbox.KeyArrowUp:
                                if cur == 0 {
                                        cur = len(index) - 1
                                        if len(index) > h {
                                                scr_y = len(index) - h + 1
                                        }
                                        break event_loop
                                } else {cur--}

                                if len(index) > h && scr_y != 0 {
                                        scr_y--
                                }

                                break event_loop
                        case tbox.KeyArrowDown:
                                if cur == len(index)-1 {
                                        cur = 0
                                        if len(index) > h {
                                                scr_y = 0
                                        }
                                        break event_loop
                                } else {cur++}

                                if len(index) > h && scr_y != len(index)-h+1 {
                                        scr_y++
                                }

                                break event_loop
                        case tbox.KeyArrowRight:
                                scr_x = scr_x + 8
                                if scr_x > dif {scr_x = dif - dif % 8}
                                break event_loop
                        case tbox.KeyArrowLeft:
                                scr_x = scr_x - 8
                                if scr_x < 0 {scr_x = 0}
                                break event_loop
                        case tbox.KeyHome:
                                cur, scr_x, scr_y = 0, 0, 0
                                break event_loop
                        case tbox.KeyEnd:
                                cur = len(index) - 1
                                if len(index) > h {
                                        scr_y = len(index)-h+1
                                }
                                break event_loop
                        }

                        switch e.Ch {
                        case 'Q':
                                break main_loop
                        case 'R':
                                chk_stat(client)
                        case 'k':
                                if cur == 0 {
                                        cur = len(index) - 1
                                        if len(index) > h {
                                                scr_y = len(index) - h + 1
                                        }
                                        break
                                } else {cur--}

                                if len(index) > h && scr_y != 0 {
                                        scr_y--
                                }
                        case 'j':
                                if cur == len(index)-1 {
                                        cur = 0
                                        if len(index) > h {
                                                scr_y = 0
                                        }
                                        break
                                } else {cur++}

                                if len(index) > h && scr_y != len(index)-h+1 {
                                        scr_y++
                                }
                        case 'l':
                                scr_x = scr_x + 8
                                if scr_x > dif {scr_x = dif - dif % 8}
                        case 'h':
                                scr_x = scr_x - 8
                                if scr_x < 0 {scr_x = 0}
                        case 'S':
                                exc("source")
                        case 'H':
                                exc("high")
                        case 'L':
                                exc("low")
                        case 'M':
                                exc("mobile")
                        case 'W':
                                exc("worst")
                        case 'A':
                                exc("audio")
                        case 'B':
                                page(PAGE_DEFAULT)
                        case 'C':
                                page(PAGE_CHAT_POPOUT)
                        case 'V':
                                page(PAGE_VIDEO_POPOUT)
                        }
                case tbox.EventResize:
                        w, h = tbox.Size()
                        cur, scr_x, scr_y = 0, 0, 0
                }

                draw_all()
        }

}

func draw_all() {
        tbox.Clear(DEF, DEF)

        for y := 0; y < len(index); y++ {
                s := fmt.Sprintf("%-*s|%-*s|%-*s",
                                c_fmt,
                                strm[index[y]],
                                g_fmt,
                                map_s_g_t[strm[index[y]]][0],
                                t_fmt,
                                map_s_g_t[strm[index[y]]][1])

                for x, c := 0, 0; x < dif ; x, c = x+1, c+1 {
                        r, k := utf8.DecodeRuneInString(s[c:])
                        if y == cur {
                                tbox.SetCell(x - scr_x, y - scr_y,
                                        r,
                                        REV,
                                        REV)
                        } else {
                                tbox.SetCell(x - scr_x, y - scr_y,
                                        r,
                                        DEF,
                                        DEF)}
                        if k != 1 {
                                for ;k != 1; k-- {
                                        c++
                                }
                        }
                }
        }

        tbox.Flush()
}

func chk_stat(client *http.Client) {
        cur, scr_x, scr_y = 0, 0, 0
        stat = make([]bool, len(strm))
        index = make([]int, 0)
        map_s_g_t = make(map[string]*[2]string, 0)
        mutex := &sync.Mutex{}
        done := make(chan bool)

        for i, s := range strm {
                go func(i int, s string) {
                        resp, err := client.Get(tapi+s+"?clientid="+cid)
                        if err != nil {log.Fatal("Error API:", err)
                        } else {
                                var m map[string]interface{}

                                js, err := ioutil.ReadAll(resp.Body)
                                if err != nil {log.Fatal("Error ReadAll:", err)}
                                err = json.Unmarshal(js, &m)
                                if err != nil {log.Fatal("Error JSON:", err)}

                                if m["stream"] != nil {
                                        stat[i] = true

                                        mutex.Lock()
                                        map_s_g_t[s] = &[2]string{
                                        m["stream"].(
                                        map[string]interface{})["game"].(
                                        string),
                                        m["stream"].(
                                        map[string]interface{})["channel"].(
                                        map[string]interface{})["status"].(
                                        string)}
                                        mutex.Unlock()
                                }
                        }
                        defer resp.Body.Close()
                        done<-true

                }(i, s)
        }

        for range strm {
                <-done
        }

        go func() {
                for i, v := range stat {
                        if v == true {index = append(index, i)}
                }
                done<-true
        }()

        g_fmt = 0
        go func() {
                for _, v := range map_s_g_t {
                        if g_fmt < len(v[0]) {g_fmt = len(v[0])}
                }
                done<-true
        }()

        t_fmt = 0
        go func() {
                for _, v := range map_s_g_t {
                        if t_fmt < len(v[1]) {t_fmt = len(v[1])}
                }
                done<-true
        }()

        for i := 0; i < 3; i++ {<-done}

        c_fmt = 0
        for _, v := range index {
                if c_fmt < len(strm[v]) {c_fmt = len(strm[v])}
        }

        dif = c_fmt + g_fmt + t_fmt + 2
}

func exc(q string) {
        if len(index) == 0 {
                tbox.Close()
                fmt.Println("No streams online.")
                os.Exit(0)
        }
        cmd := exec.Command("livestreamer", "twitch.tv/" + strm[index[cur]], q)
        cmd.Start()
}

func page(o int) {
        if len(index) == 0 {
                tbox.Close()
                fmt.Println("No streams online.")
                os.Exit(0)
        }

        var p string

        switch o {
        case PAGE_DEFAULT:
                p = "https://twitch.tv/" +
                        strm[index[cur]]
        case PAGE_CHAT_POPOUT:
                p = "https://twitch.tv/" +
                        strm[index[cur]] +
                        "/chat?popout="
        case PAGE_VIDEO_POPOUT:
                p = "https://player.twitch.tv/?channel" +
                        strm[index[cur]]
        }

        var cmd *exec.Cmd

        switch runtime.GOOS {
        case "windows":
                cmd = exec.Command("start", " ", p)
        case "darwin":
                cmd = exec.Command("open", p)
        default:
                cmd = exec.Command("xdg-open", p)
        }

        cmd.Start()
}

func parse_env(done chan bool) {
        var s string

        if runtime.GOOS == "windows" {done<-true; return
        } else {s = os.Getenv("$STRIMGO_INIT")}

        if s == "" {done<-true; return}

        i, err := strconv.Atoi(os.ExpandEnv(s))
        if err != nil {
                tbox.Close()
                log.Fatal("Error parsing env:", err)
        }

        if i != 0 {done<-true
        } else {done<-false}
}

func parse_strm(path string, done chan bool) {
	file, err := os.Open(path)
        if err != nil {
                tbox.Close()
                log.Fatal("Error opening file:", err)
        }
        defer file.Close()

        scanner := bufio.NewScanner(file)
        for scanner.Scan() {strm = append(strm, strip_string(scanner.Text()))}
        if scanner.Err() != nil {
                tbox.Close()
                log.Fatal("Error reading file:", err)
        }

        done<-true
}

func strip_string (str string) string {
        return strings.Map(func(r rune) rune {
                if r >= 32 && r != 127 {
                        return r
                }
                return -1
        }, str)
}