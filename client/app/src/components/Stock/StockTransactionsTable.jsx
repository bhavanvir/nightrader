import { useState } from "react";
import axios from "axios";

import UpArrowIcon from "../../assets/icons/UpArrowIcon";
import DownArrowIcon from "../../assets/icons/DownArrowIcon";
import TrashIcon from "../../assets/icons/TrashIcon";

// Set a default value for stockTransactions to avoid errors when it's not passed
export default function StockTransactionsTable({
  stockId,
  stockTransactions,
  showAlert,
}) {
  const [sortColumn, setSortColumn] = useState(null);
  const [sortOrder, setSortOrder] = useState("asc");

  let canadianDollar = new Intl.NumberFormat("en-CA", {
    style: "currency",
    currency: "CAD",
    minimumFractionDigits: 0,
  });

  if (!stockTransactions) {
    stockTransactions = [];
  }

  // Filter the stockTransactions to only show transactions for current stock
  stockTransactions = stockTransactions.filter(
    (stock) => stock.stock_id === stockId
  );

  const sortedTransactions = [...stockTransactions]
    .map((stock) => ({
      ...stock,
      total_cost: parseInt(stock.quantity) * parseInt(stock.stock_price),
    }))
    .sort((a, b) => {
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

  const handleClick = async (stock) => {
    await axios
      .post(
        "http://localhost/engine/cancelStockTransaction",
        {
          stock_tx_id: stock.stock_tx_id,
        },
        {
          withCredentials: true,
          headers: {
            token: localStorage.getItem("token"),
          },
        }
      )
      .then(function (response) {
        showAlert("success", "Successfully cancelled the order!");
      })
      .catch(function (error) {
        showAlert(
          "error",
          `${error.response.data.data.error}. Please try again`
        );
      });
  };

  return (
    <div className="overflow-x-auto">
      <table className="table-zebra table">
        <thead>
          <tr className="">
            <th
              className="w-1/4 text-lg"
              onClick={() => handleSort("order_status")}
            >
              <div className="flex items-center gap-2">
                Order Status{" "}
                {sortColumn === "order_status" &&
                  (sortOrder === "asc" ? <UpArrowIcon /> : <DownArrowIcon />)}
              </div>
            </th>
            <th
              className="w-1/4 text-lg"
              onClick={() => handleSort("quantity")}
            >
              <div className="flex items-center gap-2">
                Quantity{" "}
                {sortColumn === "quantity" &&
                  (sortOrder === "asc" ? <UpArrowIcon /> : <DownArrowIcon />)}
              </div>
            </th>
            <th
              className="w-1/4 text-lg"
              onClick={() => handleSort("stock_price")}
            >
              <div className="flex items-center gap-2">
                Unit Cost{" "}
                {sortColumn === "stock_price" &&
                  (sortOrder === "asc" ? <UpArrowIcon /> : <DownArrowIcon />)}
              </div>
            </th>
            <th
              className="w-1/4 text-lg"
              onClick={() => handleSort("total_cost")}
            >
              <div className="flex items-center gap-2">
                Total Cost{" "}
                {sortColumn === "total_cost" &&
                  (sortOrder === "asc" ? <UpArrowIcon /> : <DownArrowIcon />)}
              </div>
            </th>
          </tr>
        </thead>
        <tbody>
          {sortedTransactions.map((stock, index) => (
            <tr key={index}>
              <td>
                {
                  // Add a button depending on the stock.order_status is in progress display a cancel button
                  // else display nothing
                  (stock.order_status === "IN_PROGRESS") &
                  (stock.is_buy === true) ? (
                    <div className="flex justify-start gap-2 items-center">
                      {stock.order_status}
                      <button
                        className="btn"
                        onClick={() => handleClick(stock)}
                      >
                        Cancel
                        <TrashIcon />
                      </button>
                    </div>
                  ) : (
                    <div>{stock.order_status}</div>
                  )
                }
              </td>
              <td>{stock.quantity}</td>
              <td>{canadianDollar.format(stock.stock_price)}</td>
              <td>{canadianDollar.format(stock.total_cost)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
