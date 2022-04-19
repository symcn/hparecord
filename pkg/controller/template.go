package controller

import "html/template"

var indexTmpl = template.Must(template.New("index").Funcs(fn).Parse(`<html>
<head>
<title>HPA Record</title>
</head>
<style>
#endpoints {
  font-family: "Trebuchet MS", Arial, Helvetica, sans-serif;
  border-collapse: collapse;
}

#endpoints td, #endpoints th {
  border: 1px solid #ddd;
  padding: 8px;
}

#endpoints tr:nth-child(even){background-color: #f2f2f2;}

#endpoints tr:hover {background-color: #ddd;}

#endpoints th {
  padding-top: 12px;
  padding-bottom: 12px;
  text-align: left;
  background-color: #3371e3;
  color: white;
}
</style>

<body>
<br/>
<table id="endpoints" align="center" style="width:-webkit-fill-available;">
<tr>
	<th>日期</th>
	<th>集群</th>
	<th>时间</th>
	<th>事件</th>
	<th>触发消息</th>
	<th>CPU 平均使用率</th>
	<th>CPU 平均使用值</th>
</tr>
{{ noescape . }}
</table>
<br/>
</body>
</html>
`))

func noescape(str string) template.HTML {
	return template.HTML(str)
}

var fn = template.FuncMap{
	"noescape": noescape,
}
