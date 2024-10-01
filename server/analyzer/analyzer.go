package analyzer

import (
	commands "server/commands"                  
	"fmt"                       
	"os"                        
	"os/exec"                   
	"strings"                  
)


func Analyzer(inputs []string) ([]string, []string) {
    var results []string
    var errors []string

    for i, input := range inputs {

        //ignorar líneas en blanco y comentarios
        inputs := strings.TrimSpace(input)
        if inputs == "" || strings.HasPrefix(inputs, "#") {
            //continue //ignorar comentarios y líneas en blanco
            fmt.Println(inputs+"\n")
            // retornar los comentarios, para que se muestren en la consola
            results = append(results, "\n"+inputs+"\n")
            continue
        }

        tokens := strings.Fields(input)
        if len(tokens) == 0 {
            errors = append(errors, fmt.Sprintf("Comando %d: No se proporcionó ningún comando", i))
            continue
        }
        tokens[0] = strings.ToLower(tokens[0])
        var msg string
        var err error

        switch tokens[0] {
        case "mkdisk":
            _, msg, err = commands.ParserMkdisk(tokens[1:])
        case "rmdisk":
            _, msg, err = commands.ParserRmdisk(tokens[1:])
        case "fdisk":
            _, msg, err = commands.ParserFdisk(tokens[1:])
        case "mount":
            _, msg, err = commands.ParserMount(tokens[1:])
        case "mkfs":
            _, msg, err = commands.ParserMkfs(tokens[1:])
        case "rep":
            _, msg, err = commands.ParseRep(tokens[1:])
        case "list":
            _, msg, err = commands.ParseList(tokens[1:])
        case "cat":
            _, msg, err = commands.ParseCat(tokens[1:])
        case "login":
            _, msg, err = commands.ParseLogin(tokens[1:])
        case "logout":
            _, msg, err = commands.ParseLogout(tokens[1:])
        case "mkdir":
            _, msg, err = commands.ParseMkdir(tokens[1:])
        case "mkfile":
            _, msg, err = commands.ParseMkfile(tokens[1:])
        case "mkgrp":
            _, msg, err = commands.ParseMkgrp(tokens[1:])
        case "rmgrp":
            _, msg, err = commands.ParseRmgrp(tokens[1:])
        case "mkusr":
            _, msg, err = commands.ParseMkusr(tokens[1:])
        case "rmusr":
            _, msg, err = commands.ParseRmusr(tokens[1:])
        case "chgrp":
            _, msg, err = commands.ParseChgrp(tokens[1:])
        case "unmount":
            _, msg, err = commands.ParseUnmount(tokens[1:])
        case "remove":
            _, msg, err = commands.ParseRemove(tokens[1:])
        case "clear":
            cmd := exec.Command("clear")
            cmd.Stdout = os.Stdout
            err = cmd.Run()
            if err != nil {
                errors = append(errors, fmt.Sprintf("Comando %d: Error al limpiar la pantalla: %s", i, err))
            }
        default:
            err = fmt.Errorf("comando desconocido: %s", tokens[0])
        }

        if err != nil {
            errors = append(errors, err.Error())
        } else {
            results = append(results, msg)
        }
    }

    return results, errors
}
