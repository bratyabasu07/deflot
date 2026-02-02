package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	clients     = make(map[*websocket.Conn]bool)
	clientsMu   sync.RWMutex
	broadcast   = make(chan []byte, 256)
	activeScan  *exec.Cmd
	cancelScan  context.CancelFunc
	scanMu      sync.Mutex
	scanRunning bool
)

type ScanRequest struct {
	Target        string   `json:"target"`
	TargetList    string   `json:"targetList"`
	Wildcard      bool     `json:"wildcard"`
	Sources       []string `json:"sources"`
	SensitiveURLs bool     `json:"sensitiveUrls"`
	Params        bool     `json:"params"`
	JS            bool     `json:"js"`
	ExcludeLibs   bool     `json:"excludeLibs"`
	Workers       int      `json:"workers"`
	MatchCodes    []string `json:"matchCodes"`
}

type StreamMessage struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

func Start(addr string) error {
	http.HandleFunc("/", serveUI)
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/api/scan", handleScan)
	http.HandleFunc("/api/stop", handleStop)

	go handleBroadcast()

	log.Printf("[SERVER] Starting DEFLOT Web Interface on %s", addr)
	return http.ListenAndServe(addr, nil)
}

func handleBroadcast() {
	for msg := range broadcast {
		clientsMu.RLock()
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				client.Close()
				delete(clients, client)
			}
		}
		clientsMu.RUnlock()
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	defer func() {
		clientsMu.Lock()
		delete(clients, conn)
		clientsMu.Unlock()
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	scanMu.Lock()
	if scanRunning {
		scanMu.Unlock()
		http.Error(w, "Scan already running", http.StatusConflict)
		return
	}
	scanRunning = true
	scanMu.Unlock()

	var req ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		scanMu.Lock()
		scanRunning = false
		scanMu.Unlock()
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})

	go executeScan(req)
}

func handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	scanMu.Lock()
	defer scanMu.Unlock()

	if !scanRunning {
		http.Error(w, "No scan running", http.StatusBadRequest)
		return
	}

	if cancelScan != nil {
		cancelScan()
		sendMessage("system", "[ABORT] Scan termination initiated")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

func executeScan(req ScanRequest) {
	defer func() {
		scanMu.Lock()
		scanRunning = false
		activeScan = nil
		cancelScan = nil
		scanMu.Unlock()
	}()

	sendMessage("system", "[INIT] Reconnaissance engine initializing...")
	sendMessage("system", fmt.Sprintf("[TARGET] %s", req.Target))

	args := []string{"--json", "--stdout"}

	if req.Target != "" {
		args = append(args, "-d", req.Target)
	}
	if req.TargetList != "" {
		args = append(args, "-t", req.TargetList)
	}
	if req.Wildcard {
		args = append(args, "--wildcard")
		sendMessage("system", "[MODE] Wildcard subdomain enumeration enabled")
	}
	if req.SensitiveURLs {
		args = append(args, "--sensitive-urls")
		sendMessage("system", "[FILTER] Sensitive URL detection active")
	}
	if req.Params {
		args = append(args, "--params")
		sendMessage("system", "[FILTER] Parameter analysis enabled")
	}
	if req.JS {
		args = append(args, "--js")
		sendMessage("system", "[FILTER] JavaScript asset tracking active")
	}
	if req.ExcludeLibs {
		args = append(args, "--exclude-libs")
	}
	if req.Workers > 0 {
		args = append(args, "-w", fmt.Sprintf("%d", req.Workers))
		sendMessage("system", fmt.Sprintf("[WORKERS] Concurrent threads: %d", req.Workers))
	}
	if len(req.Sources) > 0 {
		var srcStr string
		for i, s := range req.Sources {
			if i > 0 {
				srcStr += ","
			}
			srcStr += s
		}
		args = append(args, "--sources", srcStr)
	}
	if len(req.MatchCodes) > 0 {
		var mcStr string
		for i, c := range req.MatchCodes {
			if i > 0 {
				mcStr += ","
			}
			mcStr += c
		}
		args = append(args, "--mc", mcStr)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancelScan = cancel

	cmd := exec.CommandContext(ctx, "deflot", args...)
	activeScan = cmd

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		sendMessage("stderr", fmt.Sprintf("[ERROR] %v", err))
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		sendMessage("stderr", fmt.Sprintf("[ERROR] %v", err))
		return
	}

	if err := cmd.Start(); err != nil {
		sendMessage("stderr", fmt.Sprintf("[ERROR] Scan initialization failed: %v", err))
		return
	}

	sendMessage("system", "[LIVE] Stream active - monitoring reconnaissance pipeline")

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			sendMessage("stdout", scanner.Text())
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			sendMessage("stderr", scanner.Text())
		}
	}()

	err = cmd.Wait()

	if ctx.Err() == context.Canceled {
		sendMessage("system", "[ABORT] Scan terminated by operator")
		sendMessage("done", "[IDLE] System ready for next operation")
	} else if err != nil {
		sendMessage("stderr", fmt.Sprintf("[ERROR] %v", err))
		sendMessage("done", "[IDLE] System ready for next operation")
	} else {
		sendMessage("system", "[COMPLETE] Reconnaissance pipeline terminated")
		sendMessage("done", "[IDLE] System ready for next operation")
	}
}

func sendMessage(msgType, content string) {
	msg := StreamMessage{
		Type:    msgType,
		Content: content,
	}
	data, _ := json.Marshal(msg)
	broadcast <- data
}

func serveUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(uiHTML))
}

const uiHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>DEFLOT | RECON OPS</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
:root{
--bg:#060608;
--panel:rgba(10,10,14,0.85);
--border:rgba(0,255,157,0.12);
--accent:#00ff9d;
--danger:#ff0044;
--warning:#ffaa00;
--cyan:#00d4ff;
--text:#e0e0e0;
--dim:#666
}
body{font-family:'Courier New',monospace;background:#060608;color:var(--text);height:100vh;overflow:hidden}
.grid{display:grid;grid-template-rows:auto 1fr;height:100vh;padding:1rem;gap:1rem}
.header{background:var(--panel);border:1px solid var(--border);padding:1.5rem;backdrop-filter:blur(15px)}
.logo{font-size:2.2rem;font-weight:700;letter-spacing:0.2em;color:var(--accent);text-shadow:0 0 30px rgba(0,255,157,0.5),0 0 60px rgba(0,255,157,0.25)}
.sub{font-size:0.7rem;color:var(--dim);letter-spacing:0.15em;margin-top:0.5rem}
.main{display:grid;grid-template-columns:300px 1fr;gap:1rem;overflow:hidden}
.panel{background:var(--panel);border:1px solid var(--border);backdrop-filter:blur(15px);overflow-y:auto}
.panel::-webkit-scrollbar{width:3px}
.panel::-webkit-scrollbar-thumb{background:var(--accent)}
.controls{padding:1.5rem}
.group{margin-bottom:1.5rem}
.label{font-size:0.65rem;color:var(--dim);letter-spacing:0.12em;text-transform:uppercase;margin-bottom:0.7rem;border-bottom:1px solid var(--border);padding-bottom:0.4rem}
input[type=text]{width:100%;background:rgba(0,0,0,0.5);border:1px solid var(--border);color:var(--text);padding:0.7rem;font-family:inherit;font-size:0.85rem;outline:none;transition:all 0.2s}
input[type=text]:focus{border-color:var(--accent);box-shadow:0 0 10px rgba(0,255,157,0.2)}
.weapons{display:grid;gap:0.6rem}
.weapon{display:flex;align-items:center;gap:0.7rem;padding:0.6rem;background:rgba(0,0,0,0.3);border:1px solid var(--border);cursor:pointer;transition:all 0.2s;user-select:none}
.weapon:hover{border-color:var(--accent);background:rgba(0,255,157,0.05)}
.weapon.active{border-color:var(--accent);background:rgba(0,255,157,0.1);box-shadow:0 0 15px rgba(0,255,157,0.15)}
.weapon input{display:none}
.indicator{width:6px;height:6px;background:var(--dim);transition:all 0.2s}
.weapon.active .indicator{background:var(--accent);box-shadow:0 0 8px var(--accent)}
.weapon-label{font-size:0.75rem;color:var(--dim);font-weight:600;letter-spacing:0.05em;transition:color 0.2s}
.weapon.active .weapon-label{color:var(--accent)}
.terminal{display:flex;flex-direction:column;position:relative}
.terminal-header{display:flex;justify-content:space-between;align-items:center;padding:1rem 1.5rem;border-bottom:1px solid var(--border)}
.status-row{display:flex;align-items:center;gap:0.7rem}
.status-dot{width:10px;height:10px;border-radius:50%;background:var(--dim);transition:all 0.3s}
.status-dot.live{background:var(--accent);box-shadow:0 0 15px var(--accent);animation:pulse 1.5s infinite}
@keyframes pulse{0%,100%{opacity:1}50%{opacity:0.6}}
.status-text{font-size:0.75rem;color:var(--dim);letter-spacing:0.1em;font-weight:600}
.status-text.live{color:var(--accent)}
.terminal-actions{display:flex;gap:0.5rem}
.tbtn{background:transparent;border:1px solid var(--border);color:var(--dim);padding:0.4rem 0.8rem;font-size:0.7rem;font-family:inherit;cursor:pointer;transition:all 0.2s;letter-spacing:0.05em}
.tbtn:hover{border-color:var(--accent);color:var(--accent)}
.output{flex:1;overflow-y:auto;padding:1.5rem;font-size:0.82rem;line-height:1.7;position:relative}
.output::-webkit-scrollbar{width:4px}
.output::-webkit-scrollbar-thumb{background:var(--border)}
.line{margin-bottom:0.15rem;white-space:pre-wrap;word-break:break-all;animation:fadeIn 0.3s}
@keyframes fadeIn{from{opacity:0;transform:translateX(-5px)}to{opacity:1;transform:translateX(0)}}
.line.system{color:var(--accent);font-weight:600}
.line.error{color:var(--danger)}
.line.js{color:var(--warning)}
.line.secret{color:var(--danger);font-weight:700}
.line.param{color:var(--cyan)}
.line.result{color:var(--accent)}
.cursor{display:inline-block;width:8px;height:14px;background:var(--accent);margin-left:2px;animation:blink 1s step-end infinite}
@keyframes blink{50%{opacity:0}}
.btn{width:100%;background:rgba(0,0,0,0.5);border:2px solid var(--accent);color:var(--accent);padding:1rem;font-family:inherit;font-size:0.9rem;font-weight:700;letter-spacing:0.15em;text-transform:uppercase;cursor:pointer;transition:all 0.3s;position:relative;overflow:hidden}
.btn::before{content:'';position:absolute;top:0;left:-100%;width:100%;height:100%;background:linear-gradient(90deg,transparent,rgba(0,255,157,0.2),transparent);transition:left 0.5s}
.btn:hover::before{left:100%}
.btn:hover{background:var(--accent);color:#000;box-shadow:0 0 25px rgba(0,255,157,0.4)}
.btn.stop{border-color:var(--danger);color:var(--danger)}
.btn.stop:hover{background:var(--danger);border-color:var(--danger);box-shadow:0 0 25px rgba(255,0,68,0.4)}
</style>
</head>
<body>
<div class="grid">
<div class="header">
<div class="logo">DEFLOT</div>
<div class="sub">RECONNAISSANCE OPERATIONS INTERFACE</div>
</div>
<div class="main">
<div class="panel controls">
<div class="group">
<div class="label">Target</div>
<input type="text" id="target" placeholder="example.com" autocomplete="off">
</div>
<div class="group">
<div class="label">Weapons</div>
<div class="weapons">
<label class="weapon" id="w-wildcard">
<input type="checkbox"><div class="indicator"></div><span class="weapon-label">WILDCARD</span>
</label>
<label class="weapon" id="w-sensitive">
<input type="checkbox"><div class="indicator"></div><span class="weapon-label">SENSITIVE</span>
</label>
<label class="weapon" id="w-params">
<input type="checkbox"><div class="indicator"></div><span class="weapon-label">PARAMS</span>
</label>
<label class="weapon" id="w-js">
<input type="checkbox"><div class="indicator"></div><span class="weapon-label">JAVASCRIPT</span>
</label>
<label class="weapon" id="w-libs">
<input type="checkbox"><div class="indicator"></div><span class="weapon-label">EXCLUDE LIBS</span>
</label>
</div>
</div>
<div class="group">
<div class="label">Workers</div>
<input type="text" id="workers" value="20" autocomplete="off">
</div>
<button class="btn" id="action-btn">START RECON</button>
</div>
<div class="panel terminal">
<div class="terminal-header">
<div class="status-row">
<div class="status-dot" id="status-dot"></div>
<div class="status-text" id="status-text">IDLE</div>
</div>
<div class="terminal-actions">
<button class="tbtn" id="clear-btn">CLEAR</button>
<button class="tbtn" id="copy-btn">COPY</button>
</div>
</div>
<div class="output" id="output"><span class="cursor"></span></div>
</div>
</div>
</div>
<script>
let ws=null,scanning=false;
const out=document.getElementById('output'),
statusDot=document.getElementById('status-dot'),
statusText=document.getElementById('status-text'),
actionBtn=document.getElementById('action-btn'),
clearBtn=document.getElementById('clear-btn'),
copyBtn=document.getElementById('copy-btn'),
targetInput=document.getElementById('target'),
workersInput=document.getElementById('workers');

const weapons={
'w-wildcard':'wildcard',
'w-sensitive':'sensitiveUrls',
'w-params':'params',
'w-js':'js',
'w-libs':'excludeLibs'
};

document.querySelectorAll('.weapon').forEach(w=>{
w.addEventListener('click',()=>{
w.classList.toggle('active');
w.querySelector('input').checked=w.classList.contains('active');
});
});

function connectWS(){
ws=new WebSocket('ws://'+location.host+'/ws');
ws.onopen=()=>{statusDot.classList.add('live');statusText.textContent='LIVE';statusText.classList.add('live')};
ws.onclose=()=>{statusDot.classList.remove('live');statusText.textContent='IDLE';statusText.classList.remove('live');setTimeout(connectWS,1000)};
ws.onmessage=(e)=>{
const msg=JSON.parse(e.data);
appendLine(msg.type,msg.content);
if(msg.type==='done'){
scanning=false;
actionBtn.textContent='START RECON';
actionBtn.classList.remove('stop');
}
};
}

function appendLine(type,content){
const cursor=out.querySelector('.cursor');
const line=document.createElement('div');
line.className='line';
if(type==='system')line.classList.add('system');
else if(type==='stderr')line.classList.add('error');
else if(type==='done')line.classList.add('system');
else{
try{
const json=JSON.parse(content);
if(json.category==='js')line.classList.add('js');
else if(json.category==='secret')line.classList.add('secret');
else if(json.category==='param')line.classList.add('param');
else line.classList.add('result');
}catch{line.classList.add('result')}
}
line.textContent=content;
if(cursor)out.insertBefore(line,cursor);
else out.appendChild(line);
out.scrollTop=out.scrollHeight;
}

actionBtn.addEventListener('click',()=>{
if(scanning){
fetch('/api/stop',{method:'POST'})
.then(res=>res.json())
.then(()=>{
scanning=false;
actionBtn.textContent='START RECON';
actionBtn.classList.remove('stop');
});
}else{
const target=targetInput.value.trim();
if(!target){appendLine('error','[ERROR] Target required');return}
const payload={
target:target,
wildcard:document.getElementById('w-wildcard').classList.contains('active'),
sensitiveUrls:document.getElementById('w-sensitive').classList.contains('active'),
params:document.getElementById('w-params').classList.contains('active'),
js:document.getElementById('w-js').classList.contains('active'),
excludeLibs:document.getElementById('w-libs').classList.contains('active'),
workers:parseInt(workersInput.value)||20
};
fetch('/api/scan',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(payload)})
.then(res=>res.json())
.then(()=>{
scanning=true;
actionBtn.textContent='STOP SCAN';
actionBtn.classList.add('stop');
})
.catch(err=>appendLine('error','[ERROR] '+err.message));
}
});

clearBtn.addEventListener('click',()=>{
const cursor=out.querySelector('.cursor');
out.innerHTML='';
if(cursor)out.appendChild(cursor);
});

copyBtn.addEventListener('click',()=>{
const text=out.textContent.replace('▌','').trim();
navigator.clipboard.writeText(text).then(()=>{
const orig=copyBtn.textContent;
copyBtn.textContent='COPIED';
setTimeout(()=>copyBtn.textContent=orig,1000);
});
});

connectWS();
</script>
</body>
</html>`
