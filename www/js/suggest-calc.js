'use strict';

(function () {
	Vue.component('v-suggest-calc-form-item',
		{
			template: `
<div>
	<div class="row d-none d-md-flex">
		<div class="col-5">
			<div class="form-check py-2 ms-4">
				<input class="form-check-input" type="checkbox" :id="inputId1" :disabled="busy" v-model="isSelected">
				<label class="form-check-label d-block" :for="inputId1">
					{{ collection.name }}
				</label>
			</div>
		</div>
		<div class="col-5">
			<input type="range" class="form-range py-4" min="0" max="100" step="5" v-model.number="value" :disabled="!isInputEnabled">
		</div>
		<div class="col-2">
			<a :href="collectionUrl" class="btn btn-primary d-block">
				<i class="bi bi-tag"></i> <span class="d-none d-xl-inline">К коллекции</span>
			</a>
		</div>
	</div>

	<div class="row d-flex d-md-none">
		<div class="col-12 d-flex">
			<div class="form-check flex-fill">
				<input class="form-check-input" type="checkbox" :id="inputId2" :disabled="busy" v-model="isSelected">
				<label class="form-check-label d-block" :for="inputId2">
					{{ collection.name }}
				</label>
			</div>
			<a :href="collectionUrl" class="btn btn-primary btn-sm d-block">
				<i class="bi bi-tag"></i> <span class="d-none d-sm-inline">К коллекции</span>
			</a>
		</div>
		<div class="col-12">
			<input type="range" class="form-range py-4" min="0" max="100" step="5" v-model.number="value" :disabled="!isInputEnabled">
		</div>
	</div>
</div>`,
			props: ['collection', 'busy'],
			computed: {
				isInputEnabled: function () {
					return !this.busy && this.isSelected;
				},
				inputId1: function () {
					return 'checkStructure_' + this.collection.id + '_2';
				},
				inputId2: function () {
					return 'checkStructure_' + this.collection.id + '_2';
				},
				collectionUrl: function(){
					return '/collections/'+ this.collection.id;
				}
			},
			data: function () {
				return {
					isSelected: false,
					value: 100
				};
			},
			methods: {
				getData: function (dict) {
					if (this.isSelected && this.value > 0) {
						dict[this.collection.id] = this.value;
					}
				}
			},
			mounted: function () {
				this.$parent.registerItem(this);
			}
		});

	Vue.component('v-suggest-calc-form',
		{
			template: `
<form class="card col-12 user-select-none" v-on:submit.prevent="submit" :disabled="busy">
	<div class="card-body">
		<h4 class="card-title">Параметры расчета</h4>
			<div class="row">
				<div class="col-12 col-md-5">
					<label for="inputAmount" class="col-form-label">Инвестируемая сумма</label>
				</div>
				<div class="col-12 col-md-7">
					<div class="input-group mb-3">
						<input type="number" id="inputAmount" class="form-control" v-bind:value="amount" :disabled="busy" autocomplete="off" min="1">
						<span class="input-group-text">&#8381;</span>
					</div>
				</div>
			</div>
			<div class="row mb-3">
				<div class="col-12 col-md-5"></div>
				<div class="col-12 col-md-7 d-flex gap-1">
					<button class="btn btn-primary btn-sm p-1 flex-fill" type="button" v-on:click="incrementAmount(10000)">
						+&nbsp;10&nbsp;000&nbsp;&#8381;
					</button>
					<button class="btn btn-primary btn-sm p-1 flex-fill" type="button" v-on:click="incrementAmount(100000)">
						+&nbsp;100&nbsp;000&nbsp;&#8381;
					</button>
					<button class="btn btn-primary btn-sm p-1 flex-fill" type="button" v-on:click="incrementAmount(1000000)">
						+&nbsp;1&nbsp;000&nbsp;000&nbsp;&#8381;
					</button>
				</div>
			</div>

			<div class="row">
				<div class="col-12 col-md-5">
					<label for="inputDuration" class="col-form-label">Срок инвестиций</label>
				</div>
				<div class="col-12 col-md-7">
					<select id="inputDuration" class="form-select" :disabled="busy" v-model="duration">
						<option v-for="d in durations" :value="d.value">{{ d.name }}</option>
					</select>
				</div>
			</div>

			<div class="row mt-4">
				<div class="col-12">
					<div class="form-check">
						<input class="form-check-input d-block" type="checkbox" id="checkStructure" v-model="enableStructure" :disabled="busy">
						<label class="form-check-label" for="checkStructure">Желаемая структура портфеля</label>
					</div>
				</div>
			</div>

			<div class="mt-4" v-if="enableStructure">
				<template v-for="collection in collections">
					<v-suggest-calc-form-item :collection="collection" :busy="busy" />
				</template>
			</div>

			<div class="row mt-4">
				<div class="col-12">
					<button type="submit" class="btn btn-primary btn-lg" :disabled="busy">
						<span v-if="!busy">
							Рассчитать портфель <i class="bi bi-arrow-right"></i>
						</span>
						<span v-if="busy">
							<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span>
							Идет расчет...
						</span>
					</button>
				</div>
			</div>
		</div>
	</div>
</form>`,
			props: ['collections'],
			data: function () {
				var data = {
					amount: 10000,
					durations: [
						{value: 1, name: 'До 1 года'},
						{value: 2, name: 'До 2 лет'},
						{value: 3, name: 'До 3 лет'},
						{value: 4, name: 'До 4 лет'},
						{value: 5, name: 'До 5 лет'},
					],
					duration: 1,
					enableStructure: false,
					items: [],
					busy: false
				};
				return data;
			},
			methods: {
				registerItem: function (item) {
					this.items.push(item);
				},

				incrementAmount: function(amount) {
					this.amount += amount;
				},

				submit() {
					var request = {
						amount: this.amount,
						max_duration: this.duration
					};

					if (this.enableStructure) {
						var dict = {};

						for (var i = 0; i < this.items.length; i++) {
							this.items[i].getData(dict)
						}
						var sum = 0;
						for (var collection in dict) {
							sum += dict[collection];
						}

						request.parts = [];
						for (var collection in dict) {
							request.parts.push({
								collection: collection,
								weight: 100 * (dict[collection] / sum)
							});
						}
					}

					this.busy = true;
					var url = '/suggest?json=' + encodeURIComponent(JSON.stringify(request));
					window.location = url;
				}
			}
		});
})(window);
