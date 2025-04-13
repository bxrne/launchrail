// explorer.js - Handles the explorer page functionality

document.addEventListener("DOMContentLoaded", function () {
  // Tab switching functionality
  const tabs = document.querySelectorAll(".tab");
  const tabContents = document.querySelectorAll(".tab-content");

  tabs.forEach((tab) => {
    tab.addEventListener("click", () => {
      // Remove active class from all tabs and contents
      tabs.forEach((t) =>
        t.classList.remove(
          "active",
          "border-bottom",
          "color-border-accent-emphasis",
          "color-fg-accent",
        ),
      );
      tabContents.forEach((tc) => tc.classList.remove("active"));

      // Add active class to the clicked tab and corresponding content
      tab.classList.add(
        "active",
        "border-bottom",
        "color-border-accent-emphasis",
        "color-fg-accent",
      );
      const target = tab.getAttribute("data-target");
      document.getElementById(target).classList.add("active");
    });
  });

  // Plot functionality
  const dataSourceSelect = document.getElementById("data-source");
  const xAxisSelect = document.getElementById("x-axis");
  const yAxisSelect = document.getElementById("y-axis");
  const zAxisSelect = document.getElementById("z-axis");
  const plotButton = document.getElementById("plot-button");
  const plotContainer = document.getElementById("plot-container");

  // Fetch JSON data instead of parsing from script tags
  const hash = window.location.pathname.split("/").pop();
  fetch(`/explore/${hash}/json`)
    .then((res) => res.json())
    .then((jsonData) => {
      headers = jsonData.headers || {};
      data = jsonData.data || {};
      // Proceed with populateColumns, etc.
      if (dataSourceSelect && headers.motion) {
        populateColumns("motion");

        // Set up event listeners
        dataSourceSelect.addEventListener("change", () => {
          populateColumns(dataSourceSelect.value);
        });

        plotButton.addEventListener("click", generatePlot);
      }
    })
    .catch((err) => {
      console.error("Error fetching explorer data:", err);
    });

  // Function to populate column selectors with options
  function populateColumnSelect(select, columns, includeEmpty = false) {
    // Clear existing options
    select.innerHTML = "";

    // Add empty option if requested
    if (includeEmpty) {
      const option = document.createElement("option");
      option.value = "";
      option.textContent = "None (2D Plot)";
      select.appendChild(option);
    }

    // Add column options if columns is an array
    if (Array.isArray(columns)) {
      columns.forEach((col) => {
        const option = document.createElement("option");
        option.value = col;
        option.textContent = col;
        select.appendChild(option);
      });
    } else {
      console.error("Columns is not an array:", columns);
    }
  }

  // Function to update column selectors based on selected data source
  function populateColumns(source) {
    const columns = headers[source];

    if (Array.isArray(columns)) {
      populateColumnSelect(xAxisSelect, columns);
      populateColumnSelect(yAxisSelect, columns);
      populateColumnSelect(zAxisSelect, columns, true);
    } else {
      console.error("Invalid columns data for source:", source);
    }
  }

  // Track whether the plot is up to date
  let isPlotUpToDate = false;

  // Function to generate plot
  function generatePlot() {
    const hash = document.getElementById("record-hash")?.value || "";
    const source = dataSourceSelect.value;
    const xAxis = xAxisSelect.value;
    const yAxis = yAxisSelect.value;
    const zAxis = zAxisSelect.value;

    if (!xAxis || !yAxis) {
      alert("Please select both X and Y axes");
      return;
    }

    fetch("/plot", {
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body: new URLSearchParams({ hash, source, xAxis, yAxis, zAxis }),
    })
      .then((resp) => resp.json())
      .then((res) => {
        if (res.error) {
          throw new Error(res.error);
        }
        Plotly.newPlot(plotContainer, res.plotData, res.plotLayout, {
          responsive: true,
          displayModeBar: true,
          modeBarButtonsToRemove: ["lasso2d", "select2d"],
        })
          .then(() => {
            // Add new download button below the plot
            const dlButton = document.createElement("button");
            dlButton.textContent = "Download Plot";
            dlButton.className = "btn btn-secondary";
            dlButton.style.display = "block";
            dlButton.style.margin = "1em auto 0 auto";
            dlButton.addEventListener("click", () => {
              const fileNameParts = [hash, xAxis, yAxis];
              if (zAxis) {
                fileNameParts.push(zAxis);
              }
              const finalFileName = fileNameParts.join("_") + ".png";
              Plotly.downloadImage(plotContainer, {
                format: "png",
                filename: finalFileName,
              });
            });
            plotContainer.appendChild(dlButton);
          })
          .catch((err) => {
            console.error("Error generating plot:", err);
            plotContainer.innerHTML = `
              <div class="flash flash-error mb-3">
                <p>Error generating plot: ${err.message}</p>
              </div>`;
          });
        // Mark plot as up to date
        isPlotUpToDate = true;
      })
      .catch((err) => {
        console.error("Error generating plot:", err);
        plotContainer.innerHTML = `
          <div class="flash flash-error mb-3">
            <p>Error generating plot: ${err.message}</p>
          </div>`;
      });
  }
});
