// simulation.js - Handles simulation response and error handling

document.addEventListener('htmx:afterSwap', function (event) {
    const responseElement = document.getElementById('response');
    const responseData = event.detail.xhr.response;

    try {
        // Parse the JSON response
        const jsonResponse = JSON.parse(responseData);

        // Clear previous response
        responseElement.innerHTML = '';

        if (jsonResponse.message) {
            // Success response
            responseElement.innerHTML = `
                <div class="p-1">
                    <div class="Toast Toast--success">
                        <span class="Toast-icon">
                            <svg width="12" height="16" viewBox="0 0 12 16" class="octicon octicon-check" aria-hidden="true">
                                <path fill-rule="evenodd" d="M12 5l-8 8-4-4 1.5-1.5L4 10l6.5-6.5L12 5z" />
                            </svg>
                        </span>
                        <span class="Toast-content">${jsonResponse.message}</span>
                    </div>
                </div>
            `;
        } else if (jsonResponse.error) {
            // Error response
            responseElement.innerHTML = `
                <div class="p-1">
                    <div class="Toast Toast--error">
                        <span class="Toast-icon">
                            <svg width="14" height="16" viewBox="0 0 14 16" class="octicon octicon-stop" aria-hidden="true">
                                <path fill-rule="evenodd" d="M10 1H4L0 5v6l4 4h6l4-4V5l-4-4zm3 9.5L9.5 14h-5L1 10.5v-5L4.5 2h5L13 5.5v5zM6 4h2v5H6V4zm0 6h2v2H6v-2z" />
                            </svg>
                        </span>
                        <span class="Toast-content">${jsonResponse.error}</span>
                    </div>
                </div>
            `;
        }
    } catch (e) {
        // Handle non-JSON responses or other errors
        console.error("Error processing response:", e);
        responseElement.innerHTML = `
            <div class="p-1">
                <div class="Toast Toast--error">
                    <span class="Toast-icon">
                        <svg width="14" height="16" viewBox="0 0 14 16" class="octicon octicon-stop" aria-hidden="true">
                            <path fill-rule="evenodd" d="M10 1H4L0 5v6l4 4h6l4-4V5l-4-4zm3 9.5L9.5 14h-5L1 10.5v-5L4.5 2h5L13 5.5v5zM6 4h2v5H6V4zm0 6h2v2H6v-2z" />
                        </svg>
                    </span>
                    <span class="Toast-content">Error processing response</span>
                </div>
            </div>
        `;
    }
});
