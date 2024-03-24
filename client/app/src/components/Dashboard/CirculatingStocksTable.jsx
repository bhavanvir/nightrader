import { useState } from "react";
import { useNavigate } from "react-router-dom";

import UpArrowIcon from "../../assets/icons/UpArrowIcon";
import DownArrowIcon from "../../assets/icons/DownArrowIcon";

// Set a default value for circulatingStocks to avoid errors when it's not passed
export default function CirculatingStocksTable({ circulatingStocks }) {
  const [sortColumn, setSortColumn] = useState(null);
  const [sortOrder, setSortOrder] = useState("asc");
  const navigate = useNavigate();

  let canadianDollar = new Intl.NumberFormat("en-CA", {
    style: "currency",
    currency: "CAD",
    minimumFractionDigits: 0,
  });

  if (!circulatingStocks) {
    circulatingStocks = [];
  }

  const sortedPortfolio = [...circulatingStocks].sort((a, b) => {
    if (sortColumn) {
      if (sortOrder === "asc") {
        return a[sortColumn] > b[sortColumn] ? 1 : -1;
      } else {
        return a[sortColumn] < b[sortColumn] ? 1 : -1;
      }
    } else {
      return 0;
    }
  });

  const handleSort = (column) => {
    if (sortColumn === column) {
      setSortOrder(sortOrder === "asc" ? "desc" : "asc");
    } else {
      setSortColumn(column);
      setSortOrder("asc");
    }
  };

  const handleStockClick = (stock) => {
    navigate(`/stock/${stock.stock_id}`, { state: { stock } });
  };

  return (
    <div className="overflow-x-auto">
      <table className="table-zebra table">
        <thead>
          <tr className="">
            <th
              className="max-w-10 text-lg"
              onClick={() => handleSort("stock_id")}
            >
              <div className="flex items-center gap-2">
                ID{" "}
                {sortColumn === "stock_id" &&
                  (sortOrder === "asc" ? <UpArrowIcon /> : <DownArrowIcon />)}
              </div>
            </th>
            <th
              className="max-w-10 text-lg"
              onClick={() => handleSort("stock_name")}
            >
              <div className="flex items-center gap-2">
                Name{" "}
                {sortColumn === "stock_name" &&
                  (sortOrder === "asc" ? <UpArrowIcon /> : <DownArrowIcon />)}
              </div>
            </th>
            <th
              className="max-w-10 text-lg"
              onClick={() => handleSort("current_price")}
            >
              <div className="flex items-center gap-2">
                Price{" "}
                {sortColumn === "current_price" &&
                  (sortOrder === "asc" ? <UpArrowIcon /> : <DownArrowIcon />)}
              </div>
            </th>
          </tr>
        </thead>
        <tbody>
          {sortedPortfolio.map((stock, index) => (
            <tr key={index}>
              <td
                onClick={() => handleStockClick(stock)}
                // Add a CSS class and style for changing cursor on hover
                style={{ cursor: "pointer" }}
              >
                {stock.stock_id}
              </td>
              <td>{stock.stock_name}</td>
              <td>{canadianDollar.format(stock.current_price)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
