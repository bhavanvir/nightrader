import { useState, useEffect } from "react";
import axios from "axios";

import StockTransactionsTable from "./StockTransactionsTable";

export default function StockTransactions({ user, Stock, showAlert }) {
  const [stockTransactions, setStockTransactions] = useState([]);

  const fetchStockTransactions = async () => {
    await axios
      .get("http://localhost/transaction/getStockTransactions", {
        withCredentials: true,
        headers: {
          token: localStorage.getItem("token"),
        },
      })
      .then(function (response) {
        setStockTransactions(response.data.data);
      })
      .catch(function (error) {
        showAlert(
          "error",
          "There was an error fetching your stock transactions. Please try again"
        );
      });
  };

  useEffect(() => {
    fetchStockTransactions();
    // If there is a stock transaction in progress, fetch the stock transactions every 5 seconds
    const intervalId = setInterval(() => {
      const isInProgress = stockTransactions.some(
        (transaction) => transaction.order_status === "IN_PROGRESS"
      );
      if (isInProgress) {
        fetchStockTransactions();
      } else {
        clearInterval(intervalId);
      }
    }, 5000);
    return () => clearInterval(intervalId);
    // eslint-disable-next-line
  }, []);

  return (
    <div className="mt-6">
      <div className="grid grid-cols-1">
        <div className="card bg-base-300 shadow-xl">
          <div className="card-body">
            <h1 className="text-xl font-bold">Stock transactions</h1>
            <StockTransactionsTable
              stockTransactions={stockTransactions}
              stockId={Stock.StockId}
              showAlert={showAlert}
            />
          </div>
        </div>
      </div>
    </div>
  );
}
