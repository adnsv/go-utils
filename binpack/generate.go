package binpack

import (
	"fmt"
	"io"
)

func appendHex(dst []byte, v byte) []byte {
	n1 := v >> 4
	if n1 >= 10 {
		n1 += 'a' - 10
	} else {
		n1 += '0'
	}
	n2 := v & 0x0f
	if n2 >= 10 {
		n2 += 'a' - 10
	} else {
		n2 += '0'
	}
	return append(dst, '0', 'x', n1, n2, ',')
}

func MakeCpp(w io.Writer, namespace string, sources []*Source) {
	fmt.Fprintf(w, "// DO NOT EDIT: Automatically-generated file\n")
	fmt.Fprintf(w, "// clang-format off\n\n")

	fmt.Fprintf(w, "#include <array>\n")
	fmt.Fprintf(w, "#include <cstddef>\n\n")

	if namespace != "" {
		fmt.Fprintf(w, "namespace %s {\n", namespace)
	}
	for _, source := range sources {
		fmt.Fprintf(w, "\nconst std::array<std::byte, %d> %s = {", len(source.Content), source.Name)
		eol := []byte("\n    ")
		buf := []byte{}
		for i, b := range source.Content {
			if i%32 == 0 {
				w.Write(buf)
				w.Write(eol)
				buf = buf[:0]
			}
			buf = appendHex(buf, b)
		}
		w.Write(buf)
		fmt.Fprintf(w, "\n};\n")
	}
	if namespace != "" {
		fmt.Fprintf(w, "\n} // namespace %s\n", namespace)
	}
}

func MakeGolang(w io.Writer, packagename string, sources []*Source) {
	fmt.Fprintf(w, "package %s\n\n", packagename)

	fmt.Fprintf(w, "// DO NOT EDIT: Automatically-generated file\n")

	for _, source := range sources {
		fmt.Fprintf(w, "\nvar %s = []byte{", source.Name)
		eol := []byte("\n    ")
		buf := []byte{}
		for i, b := range source.Content {
			if i%32 == 0 {
				w.Write(buf)
				w.Write(eol)
				buf = buf[:0]
			}
			buf = appendHex(buf, b)
		}
		w.Write(buf)
		fmt.Fprintf(w, "\n};\n")
	}
}
