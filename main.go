package main

import (
	"github.com/NDRAEY/curses"
	"golang.org/x/term"
	"log"
	"os"
	"os/signal"
	"strings"
	"errors"
	"io/ioutil"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"fmt"
)

type MinerNode struct {
	Accepted int
	Algorithm string
	Diff int
	Hashrate float32
	Identifier string
	Ki int
	Pool string
	Rejected int
	Sharetime float32
	Software string
	ThreadId string
	Username string
	Wd string
}

type Transaction struct {
	Amount float64
	Datetime string
	Hash string
	Memo string
	Id int
	Recipient string
	Sender string
}

type Data struct {
	Result struct {
		Balance struct {
			Balance float32 `json:"balance"`
			Created string
			Username string
			Verifed string
		} `json:"balance"`
		Miners []MinerNode
		Transactions []Transaction
	}
}

type Screen struct {
	Width, Height int
}

type Config struct {
	Username string
}

var data Data
var screen Screen
var config Config
var userdata Data

var configer bool = false;
var mainscreen bool = false;
//var startupanim bool = true;

var configread bool = false;
var fontread bool = false;
var windowsetup bool = false;
var firsttime bool = true;

var statusstring string = "[ ]";
var font []string;

var balancewin *curses.Window;
var statuswin *curses.Window;
var transwin *curses.Window;

var STR_SYNC string = "Synchronizing..."
var STR_FAIL_CONN string = "Failed to connect!"
var STR_OK string = "Okay! ^_^"

var timeout int

func main(){
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func(){
	    for sig := range c {
	    	sig = sig
	        curses.End()
	        print("\x1b[H\x1b[2JGoodbye!!!\n")
	        os.Exit(0)
	    }
	}()

	if _, err := os.Stat("config.json"); errors.Is(err, os.ErrNotExist) {
		configer = true;
	}else{
		mainscreen = true;
	}

	scw, sch, err := term.GetSize(0)
	if err!=nil { log.Fatal(err); os.Exit(1) }
	screen.Width, screen.Height = scw, sch

	curses.Initscr()
	defer curses.End()
	curses.Cbreak()
	curses.Noecho()
	curses.Stdscr.Keypad(true)

	curses.Stdscr.Box(0,0)
	curses.Stdscr.Mvaddstr(screen.Height-1,2,"PulseWallet by NDRAEY")
	curses.Stdscr.Refresh()

	go func(){
		for {
			key := curses.Stdscr.Getch()
			if key=='q' {
				curses.End()
		        print("\x1b[H\x1b[2JGoodbye!!!\n")
		        os.Exit(0)
			}else if key=='s'{
				timeout = 0
				continue
			}else if key=='h'{
				AlertBox("Test","Hello!",25,25)
				continue
			}
		}
	}()

	rootwin := curses.NewWindow(screen.Height-3,screen.Width-3,2,2)
	rootwin.Mvaddstr(screen.Height-5,0,"Help: q: quit; s: synchronize")
	rootwin.Refresh()

	for {
		Process(rootwin)
		rootwin.Refresh()
	}
}

func Process(window *curses.Window) {
	if configer {
		AlertBox("Welcome",
				 "           Welcome to PulseWallet!         \n"+
				 "It's a Terminal UI Wrapper for DUCO Wallet.\n"+
				 "I see, you running PulseWallet first time. \n"+
				 "\n               Let's setup!               "+
				 "\n"+
				 "\n          Press Enter to continue          ",
				 45,20)
		inp := ""
		for inp=="" { inp = InputBox(window,"What's your nickname?") }

		file, err := os.Create("config.json")
		defer file.Close()
		if err!=nil { curses.End(); log.Fatal(err) }

		err = ioutil.WriteFile("config.json",
					[]byte(
						"{\n"+
						"\t\"username\":\""+inp+"\"\n}"),
				0644)
		if err!=nil {
			curses.End()
			log.Fatal()
		}
		window.Addstr("Configuration file successfully written!\n")
		
		configer=false
	}
	
	if mainscreen {
		if !configread { // Avoid disk overload
			raw, err := ioutil.ReadFile("config.json")
			if err!=nil {
				curses.End()
				log.Fatal(err)
			}
			json.Unmarshal(raw, &config)

			configread = true
		}

		if !windowsetup { // Avoiding memory leak
			balancewin = curses.NewWindow(11,screen.Width-2,1,1)
			balancewin.Box(0,0)
			balancewin.Mvaddstr(0,3,"Overview")
			balancewin.Mvaddstr(0,16,"["+config.Username+"]")
			balancewin.Refresh()

			statuswin = curses.NewWindow(3,30,1,screen.Width-30-1)
			statuswin.Border(0,0,0,0,
			int(curses.ACS_HLINE()),
			0,0,
			int(curses.ACS_VLINE()))
			statuswin.Refresh()

			transwin = curses.NewWindow(8,screen.Width-2,12,1)
			transwin.Box(0,0)
			transwin.Mvaddstr(0,3,"Transactions (Latest)")
			transwin.Refresh()
			
			windowsetup = true
		}

		if !fontread {
			rawfont, err := ioutil.ReadFile("fonttemplate.txt")
			if err!=nil {
				curses.End()
				log.Fatal(err)
				fmt.Printf("These fonts needed by PulseMiner!")
			}

			fontsp := strings.Split(string(rawfont),"\n")
			fontfin := []string{}
			fonttmp := make([]string,6)

			for idx, elm := range fontsp {
				if idx % 6 != 0 {
					fonttmp[idx%6] = elm
				}else{
					if idx==0 {fonttmp[0] = elm; continue}
					fontfin = append(fontfin,strings.Join(fonttmp,"\n"))
					fonttmp = make([]string,6)
				}
			}
			curses.End()
			font = fontfin
			/*
			FONTMAP

			[0-9]   numbers
			10      dot
			[11-15] Letters: D, U, C, O, K
			16      Space
			*/
			fontread = true
		}
		/*
		statusstring = "[~]"
		statuswin.Mvaddstr(1,1,GenStatus(STR_SYNC,statusstring))
		statuswin.Refresh()*/
		/*
		loc, err := time.LoadLocation("Local")
		if err!=nil {
			curses.End(); log.Fatal(err); os.Exit(1)
		}
		balancewin.Mvaddstr(2,2,time.Now().In(loc).Format("02/01/2006 15:04")) // Why 15:04?
		*/
		
		if firsttime {
			bnr := GenerateBanner(0.0)
			bnrh := len(strings.Split(bnr,"\n"))
			bnrw := len(strings.Split(bnr,"\n")[0])
			DrawAt(balancewin,3,3,bnr)
			balancewin.Mvaddstr(bnrh+2,bnrw+4,"DUCO")
	
			balancewin.Refresh()
			firsttime=false
		}

		DrawBalance()
		timeout = 30
		
		for (timeout)>0 {
			time.Sleep(time.Second/2)
			timeout-=1/2
		}
	}
}

func GenStatus(st string, status string) string {
	return st+strings.Repeat(" ",28-len(st)-len(status))+status
}

func DrawBalance() {
		statusstring = "[~]"
		statuswin.Mvaddstr(1,1,GenStatus(STR_SYNC,statusstring))
		statuswin.Refresh()

		userconn, err := http.Get("https://server.duinocoin.com/users/"+config.Username)
		if err!=nil {
			timeout:=10
			for timeout>0 {
				statusstring = fmt.Sprintf("[%d]",timeout)
				statuswin.Mvaddstr(1,1,strings.Repeat(" ",18))
				statuswin.Mvaddstr(1,1,GenStatus(STR_FAIL_CONN,statusstring))
				statuswin.Refresh()
				time.Sleep(1*time.Second)
				timeout--
			}
			balancewin.Refresh()
			return
		}

		userinfo, err := ioutil.ReadAll(userconn.Body)
		if err!=nil {
			curses.End()
			log.Fatal(err)
		}
		
		json.Unmarshal(userinfo, &userdata)
		balance := userdata.Result.Balance.Balance

		bnr := GenerateBanner(balance)
		bnrh := len(strings.Split(bnr,"\n"))
		bnrw := len(strings.Split(bnr,"\n")[0])
		DrawAt(balancewin,3,3,bnr)
		balancewin.Mvaddstr(bnrh+2,bnrw+4,"DUCO")

		if len(userdata.Result.Transactions)==0 {
			transwin.Mvaddstr(1,2,"There's no transactions yet.")
		}else{
			prepdata := reverseArray(userdata.Result.Transactions)
			for idx, elm := range prepdata {
				transwin.Mvaddstr(1+idx,2,
					fmt.Sprintf("%s -> %.3f DUCO -> %s",
						elm.Sender,
						elm.Amount,
						elm.Recipient))
			}
		}
		transwin.Refresh()

		statusstring = "[OK]"
		statuswin.Mvaddstr(1,1,strings.Repeat(" ",18))
		statuswin.Mvaddstr(1,1,GenStatus(STR_OK,statusstring))
		
		statuswin.Refresh()
		balancewin.Refresh()
}

func InputBox(win *curses.Window, text string) string {
	boxw, boxh := 27, 6
	tw, th := (screen.Width-boxw)/2, (screen.Height-boxh)/2

	inp := curses.NewWindow(boxh, boxw, th, tw)
	inp.Box(0,0)
	inp.Mvaddstr(1,2,text)
	inp.Mvaddch(3,1,int(' '))
	inp.Keypad(true)

	char := byte(0)
	buffer := []byte{}
	pos := 1

	for char!=10 && char!=13 {
		char = byte(inp.Getch())
		if char==127 {
			if pos>1 {
				inp.Mvaddch(3,pos,' ')
				inp.Wmove(3,pos)
				pos--
				buffer = buffer[:pos-1]
				win.Refresh()
				inp.Refresh()
			}
			continue
		}
		inp.Addch(int(char))
		curses.Stdscr.Refresh()
		pos++
		if char==10 || char==13 {continue}
		buffer = append(buffer, char)
	}
	inp.Mvaddstr(0,0,strings.Repeat(" ",boxh*boxw))
	inp.Refresh()
	return string(buffer[:])
}

func reverseArray(input []Transaction) []Transaction {
    if len(input) == 0 {
        return input
    }
    return append(reverseArray(input[1:]), input[0]) 
}

func AlertBox(title, text string, width, height int) {
	tw, th := (screen.Width-width)/2, (screen.Height-height)/2

	alrth := curses.NewWindow(height, width, th, tw)
	alrth.Box(0,0)

	alrtt := curses.NewWindow(3,width,th,tw)
	alrtt.Border(0,0,0,0,0,0,
				int(curses.ACS_VLINE()),
				int(curses.ACS_VLINE()))
	alrtt.Mvaddstr(1,int((width-len(title))/2),title)

	alrt := curses.NewWindow(height-5,width-2,th+3,tw+1)
	alrt.Addstr(text)

	alrth.Mvaddstr(height-1,0,"")

	alrth.Refresh()
	alrt.Refresh()
	alrtt.Refresh()
	
	char := byte(0)

	for char!=10 && char!=13 {
		char = byte(alrt.Getch())
		if char==10 || char==13 {
			alrth.Mvaddstr(0,0,strings.Repeat(" ",width*height))
			alrth.Refresh()
			curses.Stdscr.Refresh()
			statuswin.Refresh()
			balancewin.Refresh()
		}
	}
}

func GenerateBanner(num float32) string {
	prefix := ""
	knum := num
	numz := 5

	if knum>1000 {
		knum/=1000
		prefix = "K"
		numz = 3
	}
	
	stri := strconv.FormatFloat(float64(knum),'f',numz,64)+prefix
	strin := make([]rune, len(stri))

	numbers := make([][]string, len(strin))
	for idx, elm := range stri {
		strin[idx] = elm
		st := 0
		if string(elm)=="." {
			st = 10
		}else if string(elm)==" " {
			st = 16
		}else if string(elm)=="K" {
			st = 15
		}else{
			st, _ = strconv.Atoi(string(elm))
		}
		numbers[idx] = strings.Split(font[st],"\n")
	}

	banner := make([]string, 6) // 5 (Font width) + 1 (Space)

	/*
	[[0,0,0,0,0],
	 [0,0,0,0,0],
	 [0,0,0,0,0],
	 [0,0,0,0,0],
	 [0,0,0,0,0]]
	*/

	for idx, _ := range banner {
		for _, nelm := range numbers {
			banner[idx] += nelm[idx]+strings.Repeat(" ",5-len(nelm[idx]))+" "
		}
	}

	return strings.Join(banner,"\n")
}

func DrawAt(win *curses.Window, y, x int, str string) {
	for idx, elm := range strings.Split(str,"\n") {
		win.Mvaddstr(y+idx,x,elm)
	}
}
