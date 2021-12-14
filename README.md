# Go bindings of [WinDivert](https://github.com/basil00/Divert)


> forked from https://github.com/imgk/divert-go

## Feature

+ Support WinDivert 2.x
+ Optional CGO support to remove dependence of WinDivert.dll, use `-tags="divert_cgo"`

More details about WinDivert please refer https://www.reqrypt.org/windivert-doc.html.

## Modification
+ Add WinDivertHelperCalcChecksums
- Remove build tag `divert_embedded` 
