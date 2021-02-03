# runtime.s
.text

# Start of the program
# (runtime/rt0_linux_amd64.s)
.global _rt0_amd64_linux
_rt0_amd64_linux:
  jmp _rt0_amd64

# (runtime/asm_amd64.s)
_rt0_amd64:
  movq 0(%rsp), %rdi # argc
  leaq 8(%rsp), %rsi # argv
  jmp runtime.rt0_go

# (runtime/asm_amd64.s)
runtime.rt0_go:
  movq %rdi, %rax # argc
  movq %rsi, %rbx # argv
  movq %rbx, runtime.__argv__+0(%rip)  # ptr
  movq %rax, runtime.__argv__+8(%rip)  # len
  movq %rax, runtime.__argv__+16(%rip) # cap

  callq runtime.heapInit
  callq runtime.argsInit # this must be after heap init
  callq runtime.__initGlobals
  callq main.__initGlobals
  callq os.init # init os.Args
  callq main.main

  movq $0, %rdi  # status 0
  movq $60, %rax # sys_exit
  syscall
# End of program

os.Exit:
  movq  8(%rsp), %rdi # arg0:status
  movq $60, %rax      # sys_exit
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

// func Open(path string, mode int, perm int) (fd int)
syscall.Open:
  movq  8(%rsp), %rax # arg0:str.ptr # @TODO This is a BUG! we should add a trailing null terminator
  movq 16(%rsp), %rdi # arg0:str.len (ignored)
  movq 24(%rsp), %rsi # arg1:flag int
  movq 32(%rsp), %rdx # arg2:perm int

  pushq %rdx # perm
  pushq %rsi # flag
  pushq %rax # path
  pushq $2   # sys_open
  callq syscall.Syscall
  addq $8 * 4, %rsp
  ret

// func Read(fd int, p []byte) (n int)
syscall.Read:
  movq  8(%rsp), %rax # arg0:fd
  movq 16(%rsp), %rdi # arg1:ptr
  movq 24(%rsp), %rsi # arg1:len (ignored)
  movq 32(%rsp), %rdx # arg1:cap
  pushq %rdx # cap
  pushq %rdi # ptr
  pushq %rax # fd
  pushq $0   # sys_read
  callq syscall.Syscall
  addq $8 * 4, %rsp
  ret

// func Write(fd int, p []byte) int
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

// func Syscall(trap, a1, a2, a3 uintptr) uintptr
syscall.Syscall:
  movq   8(%rsp), %rax # syscall number
  movq  16(%rsp), %rdi # arg0
  movq  24(%rsp), %rsi # arg1
  movq  32(%rsp), %rdx # arg2
  syscall
  ret
