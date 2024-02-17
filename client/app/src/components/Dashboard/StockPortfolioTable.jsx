import { useState } from "react";

import UpArrowIcon from "../../assets/icons/UpArrowIcon";
import DownArrowIcon from "../../assets/icons/DownArrowIcon";

export default function StockPortfolioTable({ stockPortfolio }) {
  const [sortColumn, setSortColumn] = useState(null);
  const [sortOrder, setSortOrder] = useState("asc");

  const sortedPortfolio = [...stockPortfolio].sort((a, b) => {
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

  return (
    <div className="overflow-x-auto">
      <table className="table table-zebra">
        <thead>
          <tr className="">
            <th
              className="text-lg max-w-10"
              onClick={() => handleSort("stock_id")}
            >
              <div className="flex items-center gap-2">
                Stock ID{" "}
                {sortColumn === "stock_id" &&
                  (sortOrder === "asc" ? <UpArrowIcon /> : <DownArrowIcon />)}
              </div>
            </th>
            <th
              className="text-lg max-w-10"
              onClick={() => handleSort("stock_name")}
            >
              <div className="flex items-center gap-2">
                Stock Name{" "}
                {sortColumn === "stock_name" &&
                  (sortOrder === "asc" ? <UpArrowIcon /> : <DownArrowIcon />)}
              </div>
            </th>
            <th
              className="text-lg max-w-10"
              onClick={() => handleSort("quantity_owned")}
            >
              <div className="flex items-center gap-2">
                Quantity Owned{" "}
                {sortColumn === "quantity_owned" &&
                  (sortOrder === "asc" ? <UpArrowIcon /> : <DownArrowIcon />)}
              </div>
            </th>
          </tr>
        </thead>
        <tbody>
          {sortedPortfolio.map((stock, index) => (
            <tr key={index}>
              <td>{stock.stock_id}</td>
              <td>{stock.stock_name}</td>
              <td>{stock.quantity_owned}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}