package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type cliente struct {
	conn    net.Conn
	apelido string
	bot     bool
	saida   chan string
}

type pedidoEntrada struct {
	c    *cliente
	resp chan bool
}

type mensagem struct {
	de    *cliente
	para  string
	texto string
}

type pedidoTroca struct {
	c    *cliente
	novo string
	resp chan bool
}

type servidor struct {
	entrando chan pedidoEntrada
	saindo   chan *cliente
	publicas chan mensagem
	privadas chan mensagem
	trocas   chan pedidoTroca
}

func limparApelido(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "@")
	if strings.ContainsAny(s, " \t") {
		return ""
	}
	return s
}

func enviar(c *cliente, texto string) {
	select {
	case c.saida <- texto:
	default:
	}
}

func (s *servidor) gerenciar() {
	clientes := map[string]*cliente{}

	avisarTodos := func(texto string) {
		fmt.Println(texto)
		for _, c := range clientes {
			if !c.bot {
				enviar(c, texto)
			}
		}
	}

	for {
		select {
		case p := <-s.entrando:
			if _, existe := clientes[p.c.apelido]; existe {
				p.resp <- false
				continue
			}
			clientes[p.c.apelido] = p.c
			p.resp <- true
			if p.c.bot {
				avisarTodos("Bot @" + p.c.apelido + " acabou de entrar.")
			} else {
				avisarTodos("Usuário @" + p.c.apelido + " acabou de entrar.")
			}

		case c := <-s.saindo:
			if clientes[c.apelido] == c {
				delete(clientes, c.apelido)
				close(c.saida)
				avisarTodos("Usuário @" + c.apelido + " saiu.")
			}

		case m := <-s.publicas:
			avisarTodos("@" + m.de.apelido + " disse: " + m.texto)

		case m := <-s.privadas:
			destino, ok := clientes[m.para]
			if !ok {
				enviar(m.de, "Usuário @"+m.para+" não encontrado.")
				continue
			}
			linha := "@" + m.de.apelido + " disse em privado: " + m.texto
			fmt.Println(linha)
			enviar(destino, linha)

		case t := <-s.trocas:
			if t.novo == "" {
				enviar(t.c, "Apelido inválido.")
				continue
			}
			if _, existe := clientes[t.novo]; existe {
				enviar(t.c, "O apelido @"+t.novo+" já está em uso.")
				continue
			}
			antigo := t.c.apelido
			delete(clientes, antigo)
			t.c.apelido = t.novo
			clientes[t.novo] = t.c
			avisarTodos("Usuário @" + antigo + " agora é @" + t.novo + ".")
		}
	}
}

func escritor(c *cliente) {
	for texto := range c.saida {
		fmt.Fprintln(c.conn, texto)
	}
}

func (s *servidor) atender(conn net.Conn) {
	defer conn.Close()
	leitor := bufio.NewReader(conn)
	c := &cliente{conn: conn, saida: make(chan string, 64)}

	for {
		linha, err := leitor.ReadString('\n')
		if err != nil {
			return
		}
		tipo, resto, _ := strings.Cut(strings.TrimSpace(linha), " ")
		if tipo != "ENTRAR" && tipo != "BOT" {
			fmt.Fprintln(conn, "ERRO use ENTRAR ou BOT")
			continue
		}
		nome := limparApelido(resto)
		if nome == "" {
			fmt.Fprintln(conn, "ERRO apelido inválido")
			continue
		}
		c.apelido = nome
		c.bot = tipo == "BOT"
		resp := make(chan bool, 1)
		s.entrando <- pedidoEntrada{c, resp}
		if <-resp {
			break
		}
		fmt.Fprintln(conn, "ERRO apelido já está em uso")
	}

	fmt.Fprintln(conn, "OK")
	go escritor(c)

loop:
	for {
		linha, err := leitor.ReadString('\n')
		if err != nil {
			break
		}
		tipo, resto, _ := strings.Cut(strings.TrimSpace(linha), " ")
		switch tipo {
		case "MSG":
			if resto != "" {
				s.publicas <- mensagem{de: c, texto: resto}
			}
		case "PRIV":
			para, texto, _ := strings.Cut(resto, " ")
			para = limparApelido(para)
			if para != "" && texto != "" {
				s.privadas <- mensagem{de: c, para: para, texto: texto}
			}
		case "TROCAR":
			s.trocas <- pedidoTroca{c: c, novo: limparApelido(resto)}
		case "SAIR":
			break loop
		}
	}

	s.saindo <- c
}

func main() {
	s := &servidor{
		entrando: make(chan pedidoEntrada),
		saindo:   make(chan *cliente),
		publicas: make(chan mensagem),
		privadas: make(chan mensagem),
		trocas:   make(chan pedidoTroca),
	}
	go s.gerenciar()

	listener, err := net.Listen("tcp", "localhost:4444")
	if err != nil {
		fmt.Println("Erro ao abrir a porta 4444:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Servidor ouvindo em localhost:4444")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Erro ao aceitar conexão:", err)
			continue
		}
		go s.atender(conn)
	}
}
