document.querySelector("form").addEventListener("submit", (e) => {
    e.preventDefault()

    let input = document.querySelector("input")
    const formData = new FormData();

    for (const file of input.files) {
        formData.append(file.name, file)
    }

    fetch("http://localhost:8080/upload", {
        method: "POST",
        body: formData
    }).then((resp) => {
            resp.json().then((value) => console.log(value))
        }).catch(() => console.log("failure"))
})
