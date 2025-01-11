Title: {{ .Title }}
{{- /* A comment */}}
<br>My template :-) !
<br>Some Text
{{- /* another comment deleteing newline */}}
<br>some more text
<br>{{if true}}its true {{else}} its false! {{end}}
<br>{{if isTrue }} isTrue works! {{else}} isTrue not working {{end}}
<br> Date: {{ .DateStr }}
<br> Upper: {{ "upcase-this" | upper }}
<br> Upper2: {{ .Title | upper }}