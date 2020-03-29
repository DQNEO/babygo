# runtime
.text

# Start of program
.global _start
_start:
  callq main.main
  # exit(0)
  movq $0, %rdi # arg1
  movq $60, %rax
  syscall
# End of program

runtime.printstring:
  movq $2, %rdi # stderr
  movq 16(%rsp), %rsi # arg0:buf -> arg0:buf
  movq  8(%rsp), %rdx # arg1:len -> arg1:len
  movq $1, %rax # sys_write
  syscall
  ret

os.Exit:
  movq 8(%rsp), %rdi # arg1
  movq $60, %rax
  syscall
