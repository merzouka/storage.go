let saveResponseField = document.getElementById("output")
let metadataRequestButton = document.getElementById("get-meta-data")
let metadataOutput = document.getElementById("meta-data")
let filesList = document.getElementById("files")

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
            resp.json().then((value) => {
                console.log(value);
                saveResponseField.innerHTML = ""
                saveResponseField.appendChild(document.createTextNode(JSON.stringify(value)))
            })
        }).catch(() => console.log("failure"))
})

metadataRequestButton.addEventListener("click", (_) => {
    fetch("http://localhost:8080/files").then((value) => {
        value.json().then((result) => {
            metadataOutput.innerHTML = ""
            filesList.innerHTML = ""
            metadataOutput.appendChild(document.createTextNode(JSON.stringify(result)))
            for (let file of result.result) {
                let fileElt = document.createElement("li")
                let latestLink = document.createElement("a")
                latestLink.setAttribute("href", `http://localhost:8080/files/${file.name}`)
                latestLink.appendChild(document.createTextNode(file.name))
                fileElt.appendChild(latestLink)
                const revisionList = document.createElement("ol")
                for (let revision of file.revisions) {
                    let revisionLink = document.createElement("a")
                    revisionLink.setAttribute("href", `http://localhost:8080/files/${revision.name}`.replace("#", "%23"))
                    revisionLink.appendChild(document.createTextNode(revision.name))
                    let revisionElt = document.createElement("li")
                    revisionElt.appendChild(revisionLink)
                    revisionList.appendChild(revisionElt)
                }
                fileElt.appendChild(revisionList);
                filesList.appendChild(fileElt);
            }
            console.log(filesList)
        }).catch((err) => console.log(err))
    }).catch(err => console.log(err))
})
