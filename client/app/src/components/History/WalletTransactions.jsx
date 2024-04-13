import { useState, useEffect } from "react";
import axios from "axios";

import WalletTransactionsTable from "./WalletTransactionsTable";

export default function WalletTransactions({ user, showAlert }) {
  const [walletTransactions, setWalletTransactions] = useState([]);

  const fetchWalletTransactions = async () => {
    await axios
      .get("http://localhost/transaction/getWalletTransactions", {
        withCredentials: true,
        headers: {
          token: localStorage.getItem("token"),
        },
      })
      .then(function (response) {
        setWalletTransactions(response.data.data);
      })
      .catch(function (error) {});
  };

  useEffect(() => {
    fetchWalletTransactions(); // eslint-disable-next-line
  }, []);

  return (
    <div className="card bg-base-300 shadow-xl">
      <div className="card-body">
        <h1 className="text-xl font-bold">Wallet transactions</h1>
        <WalletTransactionsTable walletTransactions={walletTransactions} />
      </div>
    </div>
  );
}
