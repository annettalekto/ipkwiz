package ipkwiz

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"golang.org/x/sys/windows"
)

const mailslotname = `\\.\mailslot\ElmehWizardMessages`
const stopatomname = `ElmehWizardStopChildProcesses`
const mailslotsize = 420

const (
	cWMTSTOP = iota
	cWMTLOGMSG
	cWMTTITLE
	cWMTSUCCESS
	cWMTERROR
	cWMTACTION
	cWMTBUTTONOK
	cWMTBUTTONERR
	cWMTCANBUSLOAD
)

//Wizard тип для отправки строк в Мастера сценариев (из них получается отчёт о проверке)
type Wizard struct {
	mailslot windows.Handle
	msgnum   uint32
}

// Init нужно вызвать, прежде чем пользоваться структурой
func (w *Wizard) Init() {
	if w == nil {
		return
	}
	w.mailslot = windows.InvalidHandle
	w.msgnum = 0
}

//Close закрыть соединение с Мастером сценариев
func (w *Wizard) Close() {
	if w == nil {
		return
	}
	if w.Opened() {
		windows.CloseHandle(w.mailslot)
		w.mailslot = windows.InvalidHandle
	}
}

//Opened показывает, открыт ли в данный момент канал связи с Мастером сценариев
func (w *Wizard) Opened() bool {
	if w == nil {
		return false
	}
	return (w.mailslot != windows.InvalidHandle)
}

// NeedStop Возвращает true, если Мастер сценариев скомандовал завершить проверку.
func (w *Wizard) NeedStop() bool {
	if nil == w || !w.Opened() {
		return true
	}

	needstop := GlobalFindAtom(stopatomname)
	if 0 != needstop {
		return true
	}
	return false
}

//Open соединиться с Мастером сценариев
func (w *Wizard) Open() bool {
	if w == nil || w.Opened() {
		return false
	}
	msn, _ := windows.UTF16PtrFromString(mailslotname)
	var err error
	w.mailslot, err = windows.CreateFile(msn,
		windows.GENERIC_WRITE,
		windows.FILE_SHARE_READ,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL, 0)

	var success bool
	success = (err == nil)

	if success {
		w.msgnum = 0
	} else {
		w.mailslot = windows.InvalidHandle
	}

	return (success)
}

// послать строку или команду в Мастер сценариев
func send(w *Wizard, msgtype uint16, s string) bool {
	if w == nil || !w.Opened() {
		return false
	}

	var written uint32
	var wizmsg bytes.Buffer

	// делаем заголовок
	binary.Write(&wizmsg, binary.LittleEndian, w.msgnum) //номер сообщения
	binary.Write(&wizmsg, binary.LittleEndian, msgtype)  //тип сообщения

	strmaxsize := mailslotsize - wizmsg.Len() // минус размер заголовка

	var err error
	var words []uint16
	words, err = windows.UTF16FromString(s) // конвертация в utf16

	if err != nil {
		return false
	}

	// если строка слишком длинная, обрезаем до максимально допустимого размера
	if len(words) > strmaxsize {
		words = words[:strmaxsize]
	}

	//добавляем строку в буфер
	for _, r := range words {
		binary.Write(&wizmsg, binary.LittleEndian, r)
	}

	//добавляем недостающие нули, чтобы длина сообщения была всегда одинаковой
	for wizmsg.Len() < mailslotsize {
		wizmsg.WriteByte(0)
	}

	// TODO: убедиться, что длина сообщения никогда не превышает mailslotsize

	// отправляем
	err = windows.WriteFile(w.mailslot, wizmsg.Bytes(), &written, nil)
	if err == nil {
		w.msgnum++
	}

	return (err == nil)
}

//LogMsg делает обычную строку в отчёте Мастера сценариев
func (w *Wizard) LogMsg(s string) {
	if w == nil {
		return
	}
	send(w, cWMTLOGMSG, s)
}

//Msg синоним для LogMsg
func (w *Wizard) Msg(s string) {
	if w == nil {
		return
	}
	w.LogMsg(s)
}

//Separator разделительная линия в логе
func (w *Wizard) Separator() {
	if w == nil {
		return
	}
	w.LogMsg(`——————————————————————————————`)
}

//Title делает строку-заголовок
func (w *Wizard) Title(s string) {
	if w == nil {
		return
	}
	send(w, cWMTTITLE, s)
}

//Success делает строку с зелёной кнопкой, символизирующей успех
func (w *Wizard) Success(s string) {
	send(w, cWMTSUCCESS, s)
}

//Error делает строку с красной кнопкой, символизирующей ошибку
func (w *Wizard) Error(s string) {
	if w == nil {
		return
	}
	send(w, cWMTERROR, s)
}

//Action меняет текст предыдущей строки в отчёте Мастера сценариев. Используется для обновления одной и той же строки во время длительных операций.
func (w *Wizard) Action(s string) {
	if w == nil {
		return
	}
	send(w, cWMTACTION, s)
}

//ButtonOk меняет текст предыдущей строки в отчёте Мастера сценариев и при этом рисует зелёную кнопку, символизирующую успех
func (w *Wizard) ButtonOk(s string) {
	if w == nil {
		return
	}
	send(w, cWMTBUTTONOK, s)
}

//ButtonErr меняет текст предыдущей строки в отчёте Мастера сценариев и при этом рисует красную кнопку, символизирующую ошибку
func (w *Wizard) ButtonErr(s string) {
	if w == nil {
		return
	}
	send(w, cWMTBUTTONERR, s)
}

//Stop сигнализирует Мастеру сценариев о том, что сценарий закончил свою работу. После этого Мастер сценариев прекращает принимать строки и формирует отчёт для печати.
func (w *Wizard) Stop() {
	if w == nil {
		return
	}
	send(w, cWMTSTOP, "")
}

//SendCanBusLoad посылает Мастеру сценариев информацию о загруженности шины CAN в процентах.
func (w *Wizard) SendCanBusLoad(percent uint8) {
	if w == nil {
		return
	}
	send(w, cWMTCANBUSLOAD, fmt.Sprintf("%d", percent))
}
