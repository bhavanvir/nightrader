import { useState } from "react";

import UpArrowIcon from "../../assets/icons/UpArrowIcon";
import DownArrowIcon from "../../assets/icons/DownArrowIcon";

// Set a default value for walletTransactions to avoid errors when it's not passed
export default function WalletTransactionsTable({ walletTransactions }) {
  const [sortColumn, setSortColumn] = useState(null);
  const [sortOrder, setSortOrder] = useState("asc");

  if (!walletTransactions) {
    walletTransactions = [];
  }

  const sortedWallet = [...walletTransactions].sort((a, b) => {
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
            <th
              className="w-1/5 text-lg"
              onClick={() => handleSort("is_debit")}
            >
              <div className="flex items-center gap-2">
                Is Debit{" "}
                {sortColumn === "is_debit" &&
                  (sortOrder === "asc" ? <UpArrowIcon /> : <DownArrowIcon />)}
              </div>
            </th>
            <th className="w-1/5 text-lg" onClick={() => handleSort("amount")}>
              <div className="flex items-center gap-2">
                Amount{" "}
                {sortColumn === "amount" &&
                  (sortOrder === "asc" ? <UpArrowIcon /> : <DownArrowIcon />)}
              </div>
            </th>
            <th
              className="w-1/5 text-lg"
              onClick={() => handleSort("wallet_tx_id")}
            >
              <div className="flex items-center gap-2">
                Wallet TX ID{" "}
                {sortColumn === "wallet_tx_id" &&
                  (sortOrder === "asc" ? <UpArrowIcon /> : <DownArrowIcon />)}
              </div>
            </th>
            <th
              className="w-1/5 text-lg"
              onClick={() => handleSort("stock_tx_id")}
            >
              <div className="flex items-center gap-2">
                Stock TX ID{" "}
                {sortColumn === "stock_tx_id" &&
                  (sortOrder === "asc" ? <UpArrowIcon /> : <DownArrowIcon />)}
              </div>
            </th>
            <th
              className="w-1/5 text-lg"
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
          {sortedWallet.map((transaction, index) => (
            <tr key={index}>
              <td>{transaction.is_debit ? "true" : "false"}</td>
              <td>{transaction.amount}</td>
              <td>{transaction.wallet_tx_id}</td>
              <td>{transaction.stock_tx_id}</td>
              <td>{transaction.time_stamp}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
