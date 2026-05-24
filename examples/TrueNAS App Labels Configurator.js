const labelsSetupFunc = async (labelContainerName, labels) => {
    const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

    const set = async (input, value) => {
        input.scrollIntoView({ block: "center" });
        input.focus();

        const setter = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, "value").set;

        setter.call(input, "");
        input.dispatchEvent(new Event("input", { bubbles: true }));

        await sleep(50);

        setter.call(input, value);
        input.dispatchEvent(new InputEvent("input", {
            bubbles: true,
            inputType: "insertText",
            data: value,
        }));
        input.dispatchEvent(new Event("change", { bubbles: true }));

        await sleep(100);
    };

    const root = document.querySelector("#labels");

    for (const [key, value] of Object.entries(labels)) {
        root.querySelector(
            ":scope > div > ix-list > div > div.label-container button[data-test='button-add-item']"
        ).click();

        await sleep(300);

        const item = [...root.querySelectorAll(
            ":scope > div > ix-list > div > div.input-container > ix-list-item"
        )].at(-1);

        await set(item.querySelector('input[aria-label="Key"]'), key);
        await set(item.querySelector('input[aria-label="Value"]'), value);

        item.querySelector('button[data-test="button-add-item-containers"]').click();

        await sleep(300);

        const containerInput = [...item.querySelectorAll('input[aria-label="Container"]')].at(-1);

        await set(containerInput, labelContainerName);

        await sleep(600);

        const option = [...document.querySelectorAll(
            ".cdk-overlay-container mat-option, .cdk-overlay-container [role='option']"
        )].find((option) => option.textContent.trim().includes(labelContainerName));

        option.click();

        console.log(`Added ${key}=${value}`);
        await sleep(200);
    }

    console.log("Done.");
};