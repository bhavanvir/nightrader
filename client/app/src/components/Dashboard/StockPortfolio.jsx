import { useState, useEffect } from "react";
import axios from "axios";

import StockPortfolioTable from "./StockPortfolioTable";

export default function StockPortfolio({ user, showAlert }) {
  const [stockPortfolio, setStockPortfolio] = useState([]);

  const fetchStockPortfolio = async () => {
    await axios
      .get("http://localhost/transaction/getStockPortfolio", {
        withCredentials: true,
        headers: {
          token: localStorage.getItem("token"),
        },
      })
      .then(function (response) {
        setStockPortfolio(response.data.data);
      })
      .catch(function (error) {
        showAlert(
          "error",
          "There was an error fetching your stock portfolio. Please try again"
        );
      });
  };

  useEffect(() => {
    fetchStockPortfolio(); // eslint-disable-next-line
  }, []);

  return (
    <div className="mt-6">
      <div className="grid grid-cols-1">
        <div className="card bg-base-300 shadow-xl">
          <div className="card-body">
            <h1 className="text-xl font-bold">Stock portfolio</h1>
            <StockPortfolioTable stockPortfolio={stockPortfolio} />
          </div>
        </div>
      </div>
    </div>
  );
}
