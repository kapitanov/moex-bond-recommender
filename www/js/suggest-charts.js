'use strict';

(function (window) {

	window.createCashFlowChart = function (el, cashflows) {
		var labels = [];
		var coupons = [];
		var amortizations = [];
		var maturities = [];
		for (var i = 0; i < cashflows.length; i++) {
			var label = new Date(cashflows[i].Date).toLocaleDateString();
			labels.push(label);
			coupons.push(cashflows[i].HasCoupon ? cashflows[i].Amount : null);
			amortizations.push(cashflows[i].HasAmortization ? cashflows[i].Amount : null);
			maturities.push(cashflows[i].HasMaturity ? cashflows[i].Amount : null);
		}

		var data = {
			labels: labels,
			datasets: [
				{
					label: 'Купоны',
					data: coupons,
					backgroundColor: 'rgb(75, 192, 192)',
					borderColor: 'rgb(75, 192, 192)'
				},
				{
					label: 'Амортизации',
					data: amortizations,
					backgroundColor: 'rgb(153, 102, 255)',
					borderColor: 'rgb(153, 102, 255)'
				},
				{
					label: 'Погашения',
					data: maturities,
					backgroundColor: 'rgb(54, 162, 235)',
					borderColor: 'rgb(54, 162, 235)'
				},
			]
		};

		var config = {
			type: 'bar',
			data: data,
			options: {}
		};

		var cashFlowChart = new Chart(document.getElementById(el), config);
	};

	window.createStructureChart = function (el, positions) {
		var labels = [];
		var values = [];
		var colors = [];
		var bgColors = [
			'rgb(255, 99, 132)',
			'rgb(255, 159, 64)',
			'rgb(255, 205, 86)',
			'rgb(75, 192, 192)',
			'rgb(54, 162, 235)',
			'rgb(153, 102, 255)',
			'rgb(201, 203, 207)'
		];
		for (var i = 0; i < positions.length; i++) {
			var label = positions[i].Bond.ShortName;
			var value = positions[i].Weight;
			labels.push(label);
			values.push(value);
			colors.push(bgColors[i % bgColors.length]);
		}

		var data = {
			labels: labels,
			datasets: [
				{
					label: '',
					data: values,
					backgroundColor: colors,
					borderColor: colors
				}
			]
		};

		var formatter = new Intl.NumberFormat({maximumSignificantDigits: 1});
		var config = {
			type: 'pie',
			data: data,
			options: {
				plugins: {
					legend: {
						display: true,
						position: 'right'
					},
					tooltip: {
						callbacks: {
							label: function (context) {
								var value = context.dataset.data[context.dataIndex];
								value *= 100;
								var label = formatter.format(value) + '%';
								return label;
							}
						}
					}
				}
			}
		};

		var cashFlowChart = new Chart(document.getElementById(el), config);
	};
})(window);
