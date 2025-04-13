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

  // Parse the data from the script tags
  let headers = {};
  let data = {};

  try {
    const headersScript = document.getElementById("headers-data");
    const dataScript = document.getElementById("table-data");

    if (headersScript && dataScript) {
      headers = JSON.parse(headersScript.textContent);
      data = JSON.parse(dataScript.textContent);
    } else {
      console.error("Data scripts not found");
      return;
    }
  } catch (e) {
    console.error("Error parsing data:", e, data, headers);
    return;
  }

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

  // Initialize the form with data from the motion source
  if (dataSourceSelect && headers.motion) {
    populateColumns("motion");

    // Set up event listeners
    dataSourceSelect.addEventListener("change", () => {
      populateColumns(dataSourceSelect.value);
    });

    plotButton.addEventListener("click", generatePlot);
  }

  // Function to generate plot
  function generatePlot() {
    const source = dataSourceSelect.value;
    const xAxis = xAxisSelect.value;
    const yAxis = yAxisSelect.value;
    const zAxis = zAxisSelect.value;

    if (!xAxis || !yAxis) {
      alert("Please select both X and Y axes");
      return;
    }

    try {
      const sourceData = data[source];
      const sourceHeaders = headers[source];

      if (!Array.isArray(sourceData) || !Array.isArray(sourceHeaders)) {
        console.error("Invalid data structure");
        return;
      }

      const xIndex = sourceHeaders.indexOf(xAxis);
      const yIndex = sourceHeaders.indexOf(yAxis);
      const zIndex = zAxis ? sourceHeaders.indexOf(zAxis) : -1;

      if (xIndex === -1 || yIndex === -1 || (zAxis && zIndex === -1)) {
        console.error("Could not find selected axes in headers");
        return;
      }

      const xData = sourceData.map((row) => parseFloat(row[xIndex]));
      const yData = sourceData.map((row) => parseFloat(row[yIndex]));
      const zData = zAxis
        ? sourceData.map((row) => parseFloat(row[zIndex]))
        : null;

      const plotLayout = {
        title: yAxis + " vs " + xAxis + (zAxis ? " vs " + zAxis : ""),

        xaxis: { title: xAxis },
        yaxis: { title: yAxis },
        margin: { t: 60, r: 40, b: 60, l: 60 },
      };

      const plotData = zAxis
        ? [
            {
              x: xData,
              y: yData,
              z: zData,
              type: "scatter3d",
              mode: "markers",
              marker: {
                size: 5,
                color: xData,
                colorscale: "Viridis",
              },
              name: `${yAxis} vs ${xAxis} vs ${zAxis}`,
            },
          ]
        : [
            {
              x: xData,
              y: yData,
              type: "scatter",
              mode: "lines+markers",
              marker: {
                size: 5,
                color: "#0366d6",
              },
              line: {
                color: "#0366d6",
              },
              name: `${yAxis} vs ${xAxis}`,
            },
          ];

      // Create the plot
      Plotly.newPlot(plotContainer, plotData, plotLayout, {
        responsive: true,
        displayModeBar: true,
        modeBarButtonsToRemove: ["lasso2d", "select2d"],
      });

      // Add a title to the plot container
      const plotTitle = document.createElement("h3");
      plotTitle.className = "mb-3 color-fg-accent";
      plotTitle.textContent = "Plot Results";

      // Only add the title if it doesn't exist already
      if (!document.querySelector("#plot-container h3")) {
        plotContainer.insertAdjacentElement("afterbegin", plotTitle);
      }
    } catch (e) {
      console.error("Error generating plot:", e);
      plotContainer.innerHTML = `
                <div class="flash flash-error mb-3">
                    <p>Error generating plot: ${e.message}</p>
                </div>
            `;
    }
  }
});
