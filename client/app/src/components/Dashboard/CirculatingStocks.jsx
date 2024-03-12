import { useState, useEffect } from "react";
import axios from "axios";

import CirculatingStocksTable from "./CirculatingStocksTable";

export default function CirculatingStocks({ user, showAlert }) {
  const [circulatingStocks, setCirculatingStocks] = useState([]);

  const fetchCirculatingStocks = async () => {
    await axios
      .get("http://localhost:5433/getStockPrices", {
        withCredentials: true,
        headers: {
          token: localStorage.getItem("token"),
        },
      })
      .then(function (response) {
        setCirculatingStocks(response.data.data);
      })
      .catch(function (error) {
        showAlert(
          "error",
          "There was an error fetching circulating stocks. Please try again",
        );
      });
  };

  useEffect(() => {
    fetchCirculatingStocks(); // eslint-disable-next-line
  }, []);

  return (
    <div className="mt-6">
      <div className="grid grid-cols-1">
        <div className="card bg-base-300 shadow-xl">
          <div className="card-body">
            <h1 className="text-xl font-bold">Stocks in circulation</h1>
            <CirculatingStocksTable circulatingStocks={circulatingStocks} />
          </div>
        </div>
      </div>
    </div>
  );
}
