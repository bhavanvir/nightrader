import { useState } from "react";

import UpArrowIcon from "../../assets/icons/UpArrowIcon";
import DownArrowIcon from "../../assets/icons/DownArrowIcon";

// Set a default value for stockTransactions to avoid errors when it's not passed
export default function StockTransactionsTable({ stockTransactions }) {
  const [sortColumn, setSortColumn] = useState(null);
  const [sortOrder, setSortOrder] = useState("asc");

  if (!stockTransactions) {
    stockTransactions = [];
  }

  const sortedTransactions = [...stockTransactions].sort((a, b) => {
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
      <table className="table-zebra table">
        <thead>
          <tr>
            <th className="w-1/8 text-lg" onClick={() => handleSort("is_buy")}>
              <div className="flex items-center gap-2">
                Is Buy{" "}
                {sortColumn === "is_buy" &&
                  (sortOrder === "asc" ? <UpArrowIcon /> : <DownArrowIcon />)}
              </div>
            </th>
            <th
              className="w-1/8 text-lg"
              onClick={() => handleSort("order_status")}
            >
              <div className="flex items-center gap-2">
                Order Status{" "}
                {sortColumn === "order_status" &&
                  (sortOrder === "asc" ? <UpArrowIcon /> : <DownArrowIcon />)}
              </div>
            </th>
            <th
              className="w-1/8 text-lg"
              onClick={() => handleSort("order_type")}
            >
              <div className="flex items-center gap-2">
                Order Type{" "}
                {sortColumn === "order_type" &&
                  (sortOrder === "asc" ? <UpArrowIcon /> : <DownArrowIcon />)}
              </div>
            </th>
            <th
              className="w-1/8 text-lg"
              onClick={() => handleSort("quantity")}
            >
              <div className="flex items-center gap-2">
                Quantity{" "}
                {sortColumn === "quantity" &&
                  (sortOrder === "asc" ? <UpArrowIcon /> : <DownArrowIcon />)}
              </div>
            </th>
            <th
              className="w-1/8 text-lg"
              onClick={() => handleSort("stock_id")}
            >
              <div className="flex items-center gap-2">
                Stock ID{" "}
                {sortColumn === "stock_id" &&
                  (sortOrder === "asc" ? <UpArrowIcon /> : <DownArrowIcon />)}
              </div>
            </th>
            <th
              className="w-1/8 text-lg"
              onClick={() => handleSort("stock_price")}
            >
              <div className="flex items-center gap-2">
                Stock Price{" "}
                {sortColumn === "stock_price" &&
                  (sortOrder === "asc" ? <UpArrowIcon /> : <DownArrowIcon />)}
              </div>
            </th>
            <th
              className="w-1/8 text-lg"
              onClick={() => handleSort("stock_tx_id")}
            >
              <div className="flex items-center gap-2">
                Stock TX ID{" "}
                {sortColumn === "stock_tx_id" &&
                  (sortOrder === "asc" ? <UpArrowIcon /> : <DownArrowIcon />)}
              </div>
            </th>
            <th
              className="w-1/8 text-lg"
              onClick={() => handleSort("time_stamp")}
            >
              <div className="flex items-center gap-2">
                Time Stamp{" "}
                {sortColumn === "time_stamp" &&
                  (sortOrder === "asc" ? <UpArrowIcon /> : <DownArrowIcon />)}
              </div>
            </th>
          </tr>
        </thead>
        <tbody>
          {sortedTransactions.map((stock, index) => (
            <tr key={index}>
              <td>{stock.is_buy ? "true" : "false"}</td>
              <td>{stock.order_status}</td>
              <td>{stock.order_type}</td>
              <td>{stock.quantity}</td>
              <td>{stock.stock_id}</td>
              <td>{stock.stock_price}</td>
              <td>{stock.stock_tx_id}</td>
              <td>{stock.time_stamp}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
