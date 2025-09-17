// document.querySelectorAll(".card-btn-add").forEach(btn => {
//     btn.addEventListener("click", async function() {
//         const userID = "1";
//         const ServiceID = btn.getAttribute('btn-id');
//         const BidID = String(await GetUserActiveBid(userID));
//         await AddSoftwareDevServiceToBid(ServiceID, BidID);
//     })
// });

// async function AddSoftwareDevServiceToBid(serviceID, bidID) {
//     try {
//         const response = await fetch("http://localhost:80/api/addservice", {
//             method: "POST",
//             headers: {
//                 'Content-Type': 'application/json',
//             },
//             body: JSON.stringify({
//                 service: serviceID,
//                 bid: bidID,
//             })
//         });

//         const data = await response.json();
//         console.log(data);
//         let bidCount = document.getElementsByClassName("extra-bid")[0];
//         bidCount.innerHTML = String(Number(bidCount.innerHTML) + 1);
//     } catch (error) {
//         console.error("error:", error);
//     }
// }