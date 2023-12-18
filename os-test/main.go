package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {

	///home/jking/vscode/system_call
	if len(os.Args) > 1 {
		err := os.Chdir(os.Args[1])
		if err != nil {
			panic(err)
		}
	}

	// ./qemu-x64/bin/qemu-system-x86_64 -kernel linux-6.2/arch/x86/boot/bzImage -boot c -m 2048M -hda rootfs-lib.img -append "root=/dev/sda rw console=ttyS0 acpi=off nokaslr" -serial stdio -display none
	cmd := exec.Command("./qemu-x64/bin/qemu-system-x86_64", "-kernel", "linux-6.2/arch/x86/boot/bzImage", "-boot", "c", "-m", "2048M", "-hda", "rootfs-lib.img", "-append", "root=/dev/sda rw console=ttyS0 acpi=off nokaslr", "-serial", "stdio", "-display", "none")
	in, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	err = cmd.Start()
	if err != nil {
		panic(err)
	}
	fmt.Println(cmd)
	time.Sleep(3 * time.Second)
	buf := make([]byte, 1024)
	output := ""
	for {
		n, err := out.Read(buf)
		if err != nil {
			panic(err)
		}
		output += string(buf[:n])
		if strings.HasSuffix(output, "Please press Enter to activate this console. ") {
			break
		}
	}
	in.Write([]byte("\n"))
	in.Write([]byte("./hello\n"))
	in.Write([]byte("exit\n"))
	time.Sleep(time.Second)

	output = ""
	for {
		n, err := out.Read(buf)
		if err != nil {
			panic(err)
		}
		output += string(buf[:n])
		if strings.Contains(output, "exit") {
			break
		}
	}
	fmt.Println(output)

	startidx := strings.Index(output, "~ # ./hello\r\n") + len("~ # ./hello\r\n")
	// 找到子串第一次出现的位置，Index一个字符
	fmt.Println(startidx)
	fmt.Println(output[startidx:])
	if strings.HasPrefix(output[startidx:], "Hello World!") {
		fmt.Println("PASS")
	} else {
		fmt.Println("FAIL")
	}

}
