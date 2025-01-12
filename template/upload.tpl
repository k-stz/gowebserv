 <!DOCTYPE html>
<html>
<title>Upload</title>
<body>

<h1>Upload a File</h1>
<p>Date: {{ .DateStr }}</p>

<form method="post" action="http://{{ .Address }}/upload-backend" enctype="multipart/form-data">
    <input type="file" name="file">
    <button>Upload</button>
</form>

</body>
</html> 