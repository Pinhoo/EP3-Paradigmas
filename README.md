# EP03 - Chat server em Go

Trabalho de um chat feito em Go com três programas separados: um servidor, um cliente e um bot. A comunicação é por TCP em `localhost:4444` (fixo, como pedido no enunciado).

## Integrantes

- João Pedro Ferreira da Rocha
- Jefferson Santana Nogueira
- Lucas de Pinho Ribeiro dos Santos

## Arquivos

- `server.go`: o servidor. Aceita várias conexões ao mesmo tempo, repassa as mensagens e trata os comandos.
- `client.go`: o cliente que o usuário usa para conversar.
- `bot.go`: o bot, que conecta como um usuário e responde as mensagens privadas com o texto invertido.

## Como rodar

Precisa só do Go instalado. Abra um terminal para cada programa, sempre começando pelo servidor:

```
go run server.go
go run bot.go
go run client.go
go run client.go
```

O bot usa o nome `Inversor` por padrão. Para trocar: `go run bot.go OutroNome`.

## Comandos do cliente

| Comando | Efeito |
|---|---|
| `\msg <mensagem>` | Mensagem pública |
| `\msg @<apelido> <mensagem>` | Mensagem privada |
| `\changenick <apelido>` | Troca o apelido |
| `\exit` | Sai do chat |

Se digitar um texto sem comando, ele é enviado como mensagem pública.

## Exemplo

No cliente:

```
\msg Olá pessoal
\msg @Inversor abcdef
\changenick GatoPreto
\exit
```

No console do servidor:

```
Bot @Inversor acabou de entrar.
Usuário @Zezinho acabou de entrar.
@Zezinho disse: Olá pessoal
@Zezinho disse em privado: abcdef
@Inversor disse em privado: fedcba
Usuário @Zezinho agora é @GatoPreto.
Usuário @GatoPreto saiu.
```

No console do bot:

```
Mensagem de: @Zezinho
Recebi: abcdef
Resposta: fedcba
```

## Como funciona

### Protocolo

Cliente e bot mandam uma linha de texto por comando e o servidor entende pela primeira palavra:

| Linha | Quem manda | Efeito |
|---|---|---|
| `ENTRAR apelido` | cliente | Registra o usuário (servidor responde `OK` ou `ERRO`) |
| `BOT apelido` | bot | Registra como bot |
| `MSG texto` | cliente | Mensagem pública |
| `PRIV apelido texto` | cliente e bot | Mensagem privada |
| `TROCAR apelido` | cliente | Troca o apelido |
| `SAIR` | cliente | Sai do chat |

### Servidor

- Cada usuário tem um canal próprio de saída (`saida`). Uma gorrotina `escritor` por conexão pega o que chega nesse canal e escreve na conexão.
- Existe uma gorrotina gerenciadora que fica num `select` esperando os eventos: entrar, sair, mensagem pública, mensagem privada e troca de apelido. Só ela mexe no mapa de usuários, então não precisei de mutex.
- Cada conexão tem uma gorrotina que lê as linhas do usuário e manda para o canal certo.
- Bots não recebem mensagens públicas nem avisos de entrada e saída, só as privadas.
- Apelidos são únicos. Se já existir, o cliente pede outro.

### Cliente

Antes de entrar, pede o apelido. Depois tem uma gorrotina que imprime o que chega do servidor e outra que lê o teclado e converte os comandos para o protocolo.

### Bot

Conecta com `BOT`, só lê mensagens no formato `@abc disse em privado: texto`, inverte o texto e devolve em privado para quem mandou.

## Apresentação

Os slides explicando a implementação estão junto com a entrega.
