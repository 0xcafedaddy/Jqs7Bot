package plugin

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jqs7/Jqs7Bot/conf"
	"github.com/jqs7/Jqs7Bot/helper"
	"github.com/jqs7/bb"
)

type Default struct{ bb.Base }

func (d *Default) Run() {
	if !d.FromGroup {
		switch d.getStatus() {
		case "auth":
			d.auth(d.Message.Text)
		case "broadcast":
			d.bc(d.Message.Text)
			d.setStatus("")
		default:
			if conf.CategoriesSet.Has(d.Message.Text) {
				// custom keyboard reply
				if !d.isAuthed() {
					d.sendQuestion()
					return
				}
				d.NewMessage(d.ChatID,
					conf.List2StringInConf(d.Message.Text)).Send()
			} else {
				if len(d.Args) > 0 {
					d.turing(d.Message.Text)
					return
				}
				photo := d.Message.Photo
				if len(photo) > 0 {
					go d.NewChatAction(d.ChatID).UploadPhoto().Send()

					fileID := photo[len(photo)-1].FileID
					link, _ := d.GetLink(fileID)
					path := helper.Downloader(link, fileID)

					mime := helper.FileMime(path)
					size := helper.FileSize(path)
					bar := helper.BarCode(path)
					vcn := helper.Vim_cn_Uploader(path)

					s := fmt.Sprintf("%s %s\n%s\n%s", mime, size, vcn, bar)
					d.NewMessage(d.ChatID, s).
						DisableWebPagePreview().
						ReplyToMessageID(d.Message.MessageID).Send()
					return
				}
			}
		}
	}
}

func (d *Default) getStatus() string {
	if conf.Redis.Exists("tgStatus:" + strconv.Itoa(d.ChatID)).Val() {
		return conf.Redis.Get("tgStatus:" + strconv.Itoa(d.ChatID)).Val()
	}
	return ""
}

func (d *Default) auth(answer string) {
	qs := conf.GetQuestions()
	index := time.Now().Hour() % len(qs)
	answer = strings.ToLower(answer)
	answer = strings.TrimSpace(answer)
	if !d.FromGroup {
		if d.isAuthed() {
			d.NewMessage(d.ChatID,
				"已经验证过了，你还想验证，你是不是傻？⊂彡☆))д`)`").
				ReplyToMessageID(d.Message.MessageID).Send()
			return
		}

		if qs[index].A.Has(answer) {
			conf.Redis.SAdd("tgAuthUser", strconv.Itoa(d.Message.From.ID))
			log.Printf("%d --- %s Auth OK\n",
				d.Message.From.ID, d.Message.From.UserName)
			d.NewMessage(d.ChatID,
				"验证成功喵~！\n原来你不是外星人呢😊").Send()
			d.setStatus("")
			d.NewMessage(d.ChatID,
				conf.List2StringInConf("help")).Send()
		} else {
			log.Printf("%d --- %s Auth Fail\n",
				d.Message.From.ID, d.Message.From.UserName)
			d.NewMessage(d.ChatID,
				"答案不对不对！你一定是外星人！不跟你玩了喵！\n"+
					"重新验证一下吧\n请问："+qs[index].Q).Send()
		}
	}
}

func (d *Default) isAuthed() bool {
	if conf.Redis.SIsMember("tgAuthUser",
		strconv.Itoa(d.Message.From.ID)).Val() {
		return true
	}
	return false
}

func (d *Default) sendQuestion() {
	if d.FromGroup {
		d.NewMessage(d.ChatID,
			"需要通过中文验证之后才能使用本功能哟~\n"+
				"点击奴家的头像进入私聊模式，进行验证吧").
			Send()
		return
	}
	qs := conf.GetQuestions()
	index := time.Now().Hour() % len(qs)
	d.NewMessage(d.ChatID,
		"需要通过中文验证之后才能使用本功能哟~\n请问："+
			qs[index].Q+"\n把答案发给奴家就可以了呢").
		Send()
	d.setStatus("auth")
}

func (d *Default) isMaster() bool {
	master := conf.GetItem("master")
	if d.Message.From.UserName == master {
		return true
	}
	return false
}

func (d *Default) setStatus(status string) {
	if status == "" {
		conf.Redis.Del("tgStatus:" +
			strconv.Itoa(d.ChatID))
		return
	}
	conf.Redis.Set("tgStatus:"+
		strconv.Itoa(d.ChatID), status, -1)
}
