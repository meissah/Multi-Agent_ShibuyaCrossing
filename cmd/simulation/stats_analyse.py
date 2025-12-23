import json
import os
import glob
import matplotlib.pyplot as plt
import numpy as np
from datetime import datetime
from pathlib import Path

class MetricsAnalyzer:
    """Analyse et compare les métriques de simulation entre Shibuya et French crossing"""
    
    def __init__(self, metrics_dir="metrics_data"):
        self.metrics_dir = metrics_dir
        self.shibuya_data = []
        self.french_data = []
        
    def load_metrics(self):
        """Charge tous les fichiers JSON de métriques"""
        json_files = glob.glob(os.path.join(self.metrics_dir, "metrics_*.json"))
        
        for filepath in json_files:
            try:
                with open(filepath, 'r', encoding='utf-8') as f:
                    data = json.load(f)
                    
                if data['CrossingType'] == 'shibuya':
                    self.shibuya_data.append(data)
                elif data['CrossingType'] == 'French':
                    self.french_data.append(data)
            except Exception as e:
                print(f"Erreur lors du chargement de {filepath}: {e}")
        
        print(f"✓ Chargé {len(self.shibuya_data)} simulations Shibuya")
        print(f"✓ Chargé {len(self.french_data)} simulations French")
    
    def calculate_averages(self, data_list):
        """Calcule les moyennes sur plusieurs simulations"""
        if not data_list:
            return None
        
        total_cars = np.mean([d['TotalCars'] for d in data_list])
        total_peds = np.mean([d['TotalPedestrians'] for d in data_list])
        avg_car_life = np.mean([d['AvgCarLifetime'] for d in data_list if d['AvgCarLifetime'] > 0])
        avg_ped_life = np.mean([d['AvgPedestrianLife'] for d in data_list if d['AvgPedestrianLife'] > 0])
        
        return {
            'total_cars': total_cars,
            'total_peds': total_peds,
            'avg_car_lifetime': avg_car_life,
            'avg_ped_lifetime': avg_ped_life,
            'sim_duration': np.mean([(datetime.fromisoformat(d['EndTime'].replace('Z', '+00:00')) - 
                                     datetime.fromisoformat(d['StartTime'].replace('Z', '+00:00'))).total_seconds() 
                                    for d in data_list])
        }
    
    def plot_comparison(self):
        """Génère les graphiques de comparaison"""
        if not self.shibuya_data and not self.french_data:
            print("Aucune donnée à analyser. Lancez d'abord des simulations.")
            return
        
        shibuya_stats = self.calculate_averages(self.shibuya_data)
        french_stats = self.calculate_averages(self.french_data)
        
        # Créer une figure avec plusieurs sous-graphiques
        fig, axes = plt.subplots(2, 2, figsize=(15, 12))
        fig.suptitle('Comparaison Shibuya vs French Crossing', fontsize=16, fontweight='bold')
        
        # 1. Temps de vie moyen (voitures)
        if shibuya_stats and french_stats:
            ax1 = axes[0, 0]
            crossing_types = ['Shibuya', 'French']
            car_lifetimes = [shibuya_stats['avg_car_lifetime'], french_stats['avg_car_lifetime']]
            bars1 = ax1.bar(crossing_types, car_lifetimes, color=['#64B5F6', '#81C784'], alpha=0.8)
            ax1.set_ylabel('Temps moyen (secondes)', fontsize=11)
            ax1.set_title('Temps de vie moyen - Voitures', fontweight='bold')
            ax1.grid(axis='y', alpha=0.3)
            
            for bar, value in zip(bars1, car_lifetimes):
                height = bar.get_height()
                ax1.text(bar.get_x() + bar.get_width()/2., height,
                        f'{value:.2f}s', ha='center', va='bottom', fontweight='bold')
        
        # 2. Temps de vie moyen (piétons)
        if shibuya_stats and french_stats:
            ax2 = axes[0, 1]
            ped_lifetimes = [shibuya_stats['avg_ped_lifetime'], french_stats['avg_ped_lifetime']]
            bars2 = ax2.bar(crossing_types, ped_lifetimes, color=['#FFB74D', '#F06292'], alpha=0.8)
            ax2.set_ylabel('Temps moyen (secondes)', fontsize=11)
            ax2.set_title('Temps de vie moyen - Piétons', fontweight='bold')
            ax2.grid(axis='y', alpha=0.3)
            
            for bar, value in zip(bars2, ped_lifetimes):
                height = bar.get_height()
                ax2.text(bar.get_x() + bar.get_width()/2., height,
                        f'{value:.2f}s', ha='center', va='bottom', fontweight='bold')
        
        # 3. Distribution des temps de vie (voitures)
        ax3 = axes[1, 0]
        if self.shibuya_data:
            shibuya_car_durations = []
            for sim in self.shibuya_data:
                shibuya_car_durations.extend([m['LifeDuration'] for m in sim['CarMetrics'] if m['LifeDuration'] > 0])
            ax3.hist(shibuya_car_durations, bins=20, alpha=0.6, label='Shibuya', color='#64B5F6')
        
        if self.french_data:
            french_car_durations = []
            for sim in self.french_data:
                french_car_durations.extend([m['LifeDuration'] for m in sim['CarMetrics'] if m['LifeDuration'] > 0])
            ax3.hist(french_car_durations, bins=20, alpha=0.6, label='French', color='#81C784')
        
        ax3.set_xlabel('Temps de vie (secondes)', fontsize=11)
        ax3.set_ylabel('Nombre de voitures', fontsize=11)
        ax3.set_title('Distribution - Voitures', fontweight='bold')
        ax3.legend()
        ax3.grid(axis='y', alpha=0.3)
        
        # 4. Distribution des temps de vie (piétons)
        ax4 = axes[1, 1]
        if self.shibuya_data:
            shibuya_ped_durations = []
            for sim in self.shibuya_data:
                shibuya_ped_durations.extend([m['LifeDuration'] for m in sim['PedestrianMetrics'] if m['LifeDuration'] > 0])
            ax4.hist(shibuya_ped_durations, bins=20, alpha=0.6, label='Shibuya', color='#FFB74D')
        
        if self.french_data:
            french_ped_durations = []
            for sim in self.french_data:
                french_ped_durations.extend([m['LifeDuration'] for m in sim['PedestrianMetrics'] if m['LifeDuration'] > 0])
            ax4.hist(french_ped_durations, bins=20, alpha=0.6, label='French', color='#F06292')
        
        ax4.set_xlabel('Temps de vie (secondes)', fontsize=11)
        ax4.set_ylabel('Nombre de piétons', fontsize=11)
        ax4.set_title('Distribution - Piétons', fontweight='bold')
        ax4.legend()
        ax4.grid(axis='y', alpha=0.3)
        
        plt.tight_layout()
        
        # Sauvegarder
        output_dir = Path(self.metrics_dir)
        timestamp = datetime.now().strftime("%Y-%m-%d_%H-%M-%S")
        output_path = output_dir / f"comparison_{timestamp}.png"
        plt.savefig(output_path, dpi=300, bbox_inches='tight')
        print(f"✓ Graphique sauvegardé: {output_path}")
        
        plt.show()
    
    def generate_report(self):
        """Génère un rapport textuel de comparaison"""
        shibuya_stats = self.calculate_averages(self.shibuya_data)
        french_stats = self.calculate_averages(self.french_data)
        
        print("\n" + "="*70)
        print("RAPPORT DE COMPARAISON - SHIBUYA vs FRENCH CROSSING")
        print("="*70)
        
        if shibuya_stats:
            print("\n📊 SHIBUYA CROSSING:")
            print(f"  Temps vie moyen voitures: {shibuya_stats['avg_car_lifetime']:.2f}s")
            print(f"  Temps vie moyen piétons:  {shibuya_stats['avg_ped_lifetime']:.2f}s")
            print(f"  Total voitures générées:  {shibuya_stats['total_cars']:.0f}")
            print(f"  Total piétons générés:    {shibuya_stats['total_peds']:.0f}")
        
        if french_stats:
            print("\n📊 FRENCH CROSSING:")
            print(f"  Temps vie moyen voitures: {french_stats['avg_car_lifetime']:.2f}s")
            print(f"  Temps vie moyen piétons:  {french_stats['avg_ped_lifetime']:.2f}s")
            print(f"  Total voitures générées:  {french_stats['total_cars']:.0f}")
            print(f"  Total piétons générés:    {french_stats['total_peds']:.0f}")
        
        if shibuya_stats and french_stats:
            print("\n🎯 ANALYSE COMPARATIVE:")
            car_diff = ((french_stats['avg_car_lifetime'] - shibuya_stats['avg_car_lifetime']) / 
                       shibuya_stats['avg_car_lifetime'] * 100)
            ped_diff = ((french_stats['avg_ped_lifetime'] - shibuya_stats['avg_ped_lifetime']) / 
                       shibuya_stats['avg_ped_lifetime'] * 100)
            
            print(f"  Différence temps voitures: {car_diff:+.1f}%")
            print(f"  Différence temps piétons:  {ped_diff:+.1f}%")
            
            if abs(car_diff) < 5 and abs(ped_diff) < 5:
                print("\n💡 CONCLUSION: Les deux systèmes sont équivalents")
            elif car_diff < 0 and ped_diff < 0:
                print("\n💡 CONCLUSION: Shibuya est plus efficace")
            elif car_diff > 0 and ped_diff > 0:
                print("\n💡 CONCLUSION: French crossing est plus efficace")
            else:
                print("\n💡 CONCLUSION: Résultats mixtes, analyse approfondie nécessaire")
        
        print("="*70 + "\n")


def main():
    """Point d'entrée principal"""
    print("🚦 Analyseur de métriques - Crossing Simulator")
    print("=" * 50)
    
    analyzer = MetricsAnalyzer()
    analyzer.load_metrics()
    
    if not analyzer.shibuya_data and not analyzer.french_data:
        print("\nAucune donnée trouvée dans 'metrics_data/'")
        print("Lancez d'abord des simulations pour générer des métriques.")
        return
    
    analyzer.generate_report()
    analyzer.plot_comparison()


if __name__ == "__main__":
    main()