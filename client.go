package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func ouvir(servidor *bufio.Reader, fim chan bool) {
	for {
		linha, err := servidor.ReadString('\n')
		if err != nil {
			break
		}
		fmt.Print(linha)
	}
	fim <- true
}

func digitar(teclado *bufio.Reader, conn net.Conn) {
	for {
		linha, err := teclado.ReadString('\n')
		if err != nil {
			fmt.Fprintln(conn, "SAIR")
			return
		}
		linha = strings.TrimSpace(linha)
		if linha == "" {
			continue
		}

		if !strings.HasPrefix(linha, "\\") {
			fmt.Fprintln(conn, "MSG "+linha)
			continue
		}

		comando, resto, _ := strings.Cut(linha, " ")
		resto = strings.TrimSpace(resto)

		switch comando {
		case "\\msg":
			if resto == "" {
				fmt.Println("Uso: \\msg <mensagem> ou \\msg @<apelido> <mensagem>")
			} else if strings.HasPrefix(resto, "@") {
				para, texto, _ := strings.Cut(resto, " ")
				texto = strings.TrimSpace(texto)
				if texto == "" {
					fmt.Println("Uso: \\msg @<apelido> <mensagem>")
				} else {
					fmt.Fprintln(conn, "PRIV "+para+" "+texto)
				}
			} else {
				fmt.Fprintln(conn, "MSG "+resto)
			}
		case "\\changenick":
			if resto == "" {
				fmt.Println("Uso: \\changenick <apelido>")
			} else {
				fmt.Fprintln(conn, "TROCAR "+resto)
			}
		case "\\exit":
			fmt.Fprintln(conn, "SAIR")
			return
		default:
			fmt.Println("Comando desconhecido:", comando)
		}
	}
}

func main() {
	conn, err := net.Dial("tcp", "localhost:4444")
	if err != nil {
		fmt.Println("Não consegui conectar no servidor:", err)
		return
	}
	defer conn.Close()

	servidor := bufio.NewReader(conn)
	teclado := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Escolha seu apelido: ")
		apelido, err := teclado.ReadString('\n')
		if err != nil {
			return
		}
		fmt.Fprintln(conn, "ENTRAR "+strings.TrimSpace(apelido))
		resp, err := servidor.ReadString('\n')
		if err != nil {
			fmt.Println("Conexão encerrada pelo servidor.")
			return
		}
		resp = strings.TrimSpace(resp)
		if resp == "OK" {
			break
		}
		fmt.Println(strings.TrimPrefix(resp, "ERRO "))
	}

	fmt.Println("Conectado! Comandos: \\msg, \\msg @apelido, \\changenick, \\exit")

	fim := make(chan bool)
	go ouvir(servidor, fim)
	go digitar(teclado, conn)
	<-fim
}
