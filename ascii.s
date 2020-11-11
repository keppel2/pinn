.data
msg:
   .ascii "arm64"
.text
.global main
main:
  mov x0, 1
  ldr x1, =msg
  mov x2, 5
  mov x8, 64
  svc 0
  ret
