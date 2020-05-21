# runtime
.text

# Start of program
.global _start
_start:
  callq runtime.heapInit
  callq main.main

  movq $0, %rdi # arg1
  movq $60, %rax # sys exit
  syscall
# End of program

os.Exit:
  movq  8(%rsp), %rdi # arg0:status
  movq $60, %rax # sys exit
  syscall

runtime.printstring:
  movq  8(%rsp), %rdi # arg0:ptr
  movq 16(%rsp), %rsi # arg1:len
  pushq %rsi
  pushq %rdi
  pushq $2 # stderr
  pushq $1 # sys_write
  callq syscall.Syscall
  addq $8 * 4, %rsp
  ret

syscall.Open:
  movq  8(%rsp), %rax # arg0:str.ptr
  movq 16(%rsp), %rdi # arg0:str.len (throw away)
  movq 24(%rsp), %rsi # arg1:flag int
  movq 32(%rsp), %rdx # arg2:perm int

  pushq %rdx # perm
  pushq %rsi # flag
  pushq %rax # path
  pushq $2   # sys_open
  callq syscall.Syscall
  addq $8 * 4, %rsp
  ret

syscall.Read:
  movq  8(%rsp), %rax # arg0:fd
  movq 16(%rsp), %rdi # arg1:ptr
  movq 24(%rsp), %rsi # arg1:len (throw away)
  movq 32(%rsp), %rdx # arg1:cap
  pushq %rdx # cap
  pushq %rdi # ptr
  pushq %rax # fd
  pushq $0   # sys_read
  callq syscall.Syscall
  addq $8 * 4, %rsp
  ret

syscall.Write:
  movq  8(%rsp), %rax # arg0:fd
  movq 16(%rsp), %rdi # arg1:ptr
  movq 24(%rsp), %rsi # arg2:len
  pushq %rsi
  pushq %rdi
  pushq %rax
  pushq $1  # sys_write
  callq syscall.Syscall
  addq $8 * 4, %rsp
  ret

syscall.Syscall:
  movq   8(%rsp), %rax # trap
  movq  16(%rsp), %rdi # arg0
  movq  24(%rsp), %rsi # arg1
  movq  32(%rsp), %rdx # arg2
  syscall
  ret
