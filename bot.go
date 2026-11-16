package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func inverter(texto string) string {
	letras := []rune(texto)
	for i, j := 0, len(letras)-1; i < j; i, j = i+1, j-1 {
		letras[i], letras[j] = letras[j], letras[i]
	}
	return string(letras)
}

func main() {
	nome := "Inversor"
	if len(os.Args) > 1 {
		nome = os.Args[1]
	}

	conn, err := net.Dial("tcp", "localhost:4444")
	if err != nil {
		fmt.Println("Não consegui conectar no servidor:", err)
		return
	}
	defer conn.Close()

	servidor := bufio.NewReader(conn)
	fmt.Fprintln(conn, "BOT "+nome)
	resp, err := servidor.ReadString('\n')
	if err != nil || strings.TrimSpace(resp) != "OK" {
		fmt.Println("O servidor não aceitou o bot:", strings.TrimSpace(resp))
		return
	}
	fmt.Println("Bot @" + nome + " conectado, esperando mensagens privadas")

	for {
		linha, err := servidor.ReadString('\n')
		if err != nil {
			fmt.Println("Conexão encerrada pelo servidor.")
			return
		}
		linha = strings.TrimSpace(linha)

		de, texto, achou := strings.Cut(linha, " disse em privado: ")
		if !achou || !strings.HasPrefix(de, "@") {
			continue
		}

		resposta := inverter(texto)
		fmt.Println("Mensagem de:", de)
		fmt.Println("Recebi:", texto)
		fmt.Println("Resposta:", resposta)
		fmt.Fprintln(conn, "PRIV "+de+" "+resposta)
	}
}
