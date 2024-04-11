/* eslint eqeqeq: 0 */

import React, { useState } from "react";
import axios from "axios";

import SellCard from "../../assets/icons/SellCard";

export default function SellTrigger({ Stock, showAlert }) {
  const [trigger, setTrigger] = useState("");
  const [quantity, setQuantity] = useState(0);
  const [price, setPrice] = useState(0);

  const handleClick = async () => {
    await axios
      .post(
        "http://localhost/engine/placeStockOrder",
        {
          stock_id: Stock.StockId,
          is_buy: false,
          order_type: trigger.toUpperCase(),
          quantity: parseInt(quantity),
          price: trigger === "Market" ? null : parseInt(price),
        },
        {
          withCredentials: true,
          headers: {
            token: localStorage.getItem("token"),
          },
        }
      )
      .then(function (response) {
        showAlert(
          "success",
          `Successfully set a ${trigger.toLowerCase()} sell order!`
        );
      })
      .catch(function (error) {
        showAlert(
          "error",
          `${error.response.data.data.error}. Please try again`
        );
      });
  };

  return (
    <div className="mt-6">
      <div className="grid grid-cols-1">
        <div className="card bg-base-300 shadow-xl">
          <div className="card-body">
            <div className="flex items-center justify-between">
              <h1 className="text-xl font-bold">Sell triggers</h1>
              <button
                className="btn"
                onClick={() =>
                  document.getElementById("sell-trigger-modal").showModal()
                }
              >
                Sell
                <SellCard />
              </button>
            </div>

            <dialog id="sell-trigger-modal" className="modal">
              <div className="modal-box w-auto h-auto overflow-hidden">
                <div>
                  <h3 className="text-lg font-bold pb-4">
                    What type of trigger would you like to set?
                  </h3>
                  <select
                    className="select select-bordered w-full max-w-xs"
                    onChange={(event) => setTrigger(event.target.value)}
                    defaultValue={"DEFAULT"}
                  >
                    <option value="DEFAULT" disabled>
                      Select a trigger
                    </option>
                    <option>Market</option>
                    <option>Limit</option>
                  </select>
                </div>

                <div className="pt-8">
                  <h3 className="text-lg font-bold pb-4">
                    How many shares would you like to sell?
                  </h3>
                  <input
                    id="quantity-modal-input"
                    type="number"
                    placeholder="Enter an amount"
                    className="input input-bordered w-full max-w-xs"
                    onKeyPress={(event) => {
                      // Prevent non-numeric characters except for the negative sign
                      if (!/[0-9]/.test(event.key)) {
                        event.preventDefault();
                      } else if (event.key === "-") {
                        // Disable the button if the input is negative
                        document.getElementById(
                          "quantity-modal-input"
                        ).disabled = true;
                      }
                    }}
                    onChange={(event) => {
                      // Prevent negative values from being entered
                      if (event.target.value < 0) {
                        event.target.value = 0;
                      }
                      setQuantity(event.target.value);
                    }}
                  />
                </div>

                <div className="pt-8">
                  <h3 className="text-lg font-bold pb-4">
                    At what price would you like to sell?
                  </h3>
                  <input
                    id="price-modal-input"
                    type="number"
                    placeholder="Enter an amount"
                    className="input input-bordered w-full max-w-xs"
                    onKeyPress={(event) => {
                      if (!/[0-9]/.test(event.key)) {
                        event.preventDefault();
                      } else if (event.key === "-") {
                        document.getElementById(
                          "price-modal-input"
                        ).disabled = true;
                      }
                    }}
                    onChange={(event) => {
                      if (event.target.value < 0) {
                        event.target.value = 0;
                      }
                      setPrice(event.target.value);
                    }}
                    disabled={trigger === "Market"}
                  />
                </div>

                <div className="pt-8 flex justify-start">
                  <button
                    className="btn"
                    onClick={handleClick}
                    id="purchase-button"
                    disabled={
                      // eslint-disable-next-line
                      (trigger === "Market" && (quantity == 0 || !quantity)) ||
                      (trigger === "Limit" &&
                        (quantity == 0 || !quantity || price == 0 || !price)) ||
                      (trigger === "" &&
                        (quantity == 0 || !quantity) &&
                        (price == 0 || !price))
                    }
                  >
                    Make purchase
                  </button>
                </div>
              </div>
              <form method="dialog" className="modal-backdrop">
                <button>close</button>
              </form>
            </dialog>
          </div>
        </div>
      </div>
    </div>
  );
}
