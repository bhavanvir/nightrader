import React, { useState } from "react";
import { useLocation } from "react-router-dom";

import Header from "../Header/Header";
import BuyTrigger from "./BuyTrigger";
import SellTrigger from "./SellTrigger";
import StockTransactions from "./StockTransactions";

const Stock = ({ stock, user }) => {
  let canadianDollar = new Intl.NumberFormat("en-CA", {
    style: "currency",
    currency: "CAD",
    minimumFractionDigits: 0,
  });

  const state = useLocation();
  const { stock_id, stock_name, current_price } = state.state.stock;
  const Stock = {
    StockId: stock_id,
    StockName: stock_name,
    CurrentPrice: current_price,
  };

  const [showAlert, setShowAlert] = useState(false);
  const [alertType, setAlertType] = useState("");
  const [message, setMessage] = useState("");

  if (!user.user_name) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="text-center">
          <h1 className="text-xl font-bold">
            Refresh your page to continue where you're going!
          </h1>
          <span className="loading loading-dots loading-lg text-primary" />
        </div>
      </div>
    );
  }

  const showAlertMessage = (type, message) => {
    setAlertType(type);
    setMessage(message);
    setShowAlert(true);
    setTimeout(() => {
      setShowAlert(false);
      setAlertType("");
      setMessage("");
    }, 5000);
  };

  return (
    <div>
      {showAlert && (
        <div role="alert" className={`alert alert-${alertType}`}>
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="h-6 w-6 shrink-0 stroke-current"
            fill="none"
            viewBox="0 0 24 24"
          >
            {alertType === "success" ? (
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            ) : (
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            )}
          </svg>
          <span>{message}</span>
        </div>
      )}
      <Header user={user} showAlert={showAlertMessage} />
      <div>
        <div className="grid grid-rows-2 justify-center pt-12">
          <div className="font-bold text-4xl text-center">
            {Stock.StockName}
          </div>
          <div className="text-2xl text-center pt-2">
            {canadianDollar.format(Stock.CurrentPrice)}
          </div>
        </div>

        <div className="container mx-auto w-[75rem]">
          <div className="grid grid-cols-2 gap-6">
            <BuyTrigger Stock={Stock} showAlert={showAlertMessage} />
            <SellTrigger Stock={Stock} showAlert={showAlertMessage} />
          </div>
          <StockTransactions
            user={user}
            Stock={Stock}
            showAlert={showAlertMessage}
          />
        </div>
      </div>
    </div>
  );
};

export default Stock;
