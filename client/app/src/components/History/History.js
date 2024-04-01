import React from "react";

import Header from "../Header/Header";
import StockTransactions from "./StockTransactions";
import WalletTransactions from "./WalletTransactions";

const History = ({ user }) => {
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

  return (
    <div>
      <Header user={user} />
      <div className="container mx-auto pt-12">
        <div className="grid grid-rows-2 gap-6">
          <StockTransactions user={user} />
          <WalletTransactions user={user} />
        </div>
      </div>
    </div>
  );
};

export default History;
