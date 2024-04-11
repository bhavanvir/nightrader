import { useState, useEffect } from "react";
import axios from "axios";

import StockTransactionsTable from "./StockTransactionsTable";

export default function StockTransactions({ user, showAlert }) {
  const [stockTransactions, setStockTransactions] = useState([]);

  const fetchStockPortfolio = async () => {
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
      .catch(function (error) {});
  };

  useEffect(() => {
    fetchStockPortfolio(); // eslint-disable-next-line
  }, []);

  return (
    <div className="card bg-base-300 shadow-xl">
      <div className="card-body">
        <h1 className="text-xl font-bold">Stock transactions</h1>
        <StockTransactionsTable stockTransactions={stockTransactions} />
      </div>
    </div>
  );
}
